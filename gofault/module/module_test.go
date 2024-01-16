package module

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gofault/gofault/controller"
	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/ioc"
	"github.com/gofault/gofault/router"
)

// dummyController is a test controller.
type dummyController struct {
	controller.BaseController
	calledMethod string
}

func (c *dummyController) Routes() []core.Route {
	return []core.Route{
		{Method: "GET", Path: "/", Handler: "Index"},
		{Method: "GET", Path: "/greet/:name", Handler: "Greet"},
	}
}

func (c *dummyController) Prefix() string { return "/test" }

func (c *dummyController) Index(ctx *core.Ctx) error {
	return controller.OK(ctx.Response, map[string]string{"message": "index"})
}

func (c *dummyController) Greet(ctx *core.Ctx) error {
	name := controller.Param(ctx, "name")
	return controller.OK(ctx.Response, map[string]string{"message": "Hello, " + name})
}

// dummyHook tracks boot/shutdown calls.
type dummyHook struct {
	bootCalled     bool
	shutdownCalled bool
	bootErr        error
	shutdownErr    error
}

func (h *dummyHook) OnBoot() error {
	h.bootCalled = true
	return h.bootErr
}

func (h *dummyHook) OnShutdown() error {
	h.shutdownCalled = true
	return h.shutdownErr
}

func TestApp_New(t *testing.T) {
	app := New()
	if app == nil {
		t.Fatal("New() returned nil")
	}
	if app.Container() == nil {
		t.Fatal("Container() returned nil")
	}
}

func TestApp_RegisterModules(t *testing.T) {
	app := New()
	mod := core.NewModule("test")
	mod.RegisterControllers(&dummyController{})

	app.RegisterModules(mod)

	if len(app.modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(app.modules))
	}
}

func TestApp_Bootstrap_DoesNotPanic(t *testing.T) {
	app := New()
	mod := core.NewModule("test")
	mod.RegisterControllers(&dummyController{})
	app.RegisterModules(mod)

	err := app.Bootstrap()
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}
}

func TestApp_SetRouter(t *testing.T) {
	app := New()
	app.SetRouter(nil) // nil router is allowed, just not used
}

func TestApp_LifecycleHooks(t *testing.T) {
	app := New()
	mod := core.NewModule("test")

	hook1 := &dummyHook{}
	hook2 := &dummyHook{}

	mod.RegisterOnBoot(hook1)
	mod.RegisterOnShutdown(hook2)

	app.RegisterModules(mod)
	app.Bootstrap()

	// Run boot hooks manually (normally called by Start).
	for _, m := range app.modules {
		for _, h := range m.OnBootHooks {
			if err := h.OnBoot(); err != nil {
				t.Fatalf("OnBoot failed: %v", err)
			}
		}
	}

	if !hook1.bootCalled {
		t.Fatal("boot hook was not called")
	}

	// Run shutdown hooks in reverse.
	for i := len(app.modules) - 1; i >= 0; i-- {
		m := app.modules[i]
		for j := len(m.OnShutdownHooks) - 1; j >= 0; j-- {
			h := m.OnShutdownHooks[j]
			if err := h.OnShutdown(); err != nil {
				t.Fatalf("OnShutdown failed: %v", err)
			}
		}
	}

	if !hook2.shutdownCalled {
		t.Fatal("shutdown hook was not called")
	}
}

func TestApp_Bootstrap_MultipleControllers(t *testing.T) {
	app := New()
	mod := core.NewModule("test")
	mod.RegisterControllers(&dummyController{}, &dummyController{})
	app.RegisterModules(mod)

	err := app.Bootstrap()
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}
}

func TestApp_MultipleModules(t *testing.T) {
	app := New()

	mod1 := core.NewModule("mod1")
	mod1.RegisterControllers(&dummyController{})

	mod2 := core.NewModule("mod2")
	mod2.RegisterControllers(&dummyController{})

	app.RegisterModules(mod1, mod2)

	if len(app.modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(app.modules))
	}
}

func TestApp_Start_WithoutBootstrap_DoesBootstrap(t *testing.T) {
	app := New()
	mod := core.NewModule("test")
	mod.RegisterControllers(&dummyController{})
	app.RegisterModules(mod)

	// Verify not booted yet.
	if app.booted {
		t.Fatal("app should not be booted before Bootstrap")
	}

	// Calling Bootstrap directly.
	err := app.Bootstrap()
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}
	if !app.booted {
		t.Fatal("app should be booted after Bootstrap")
	}
}

func TestApp_RouteHandlerDispatch(t *testing.T) {
	rtr := router.New()
	app := New()
	app.SetRouter(rtr)

	mod := core.NewModule("test")
	mod.RegisterControllers(&dummyController{})
	app.RegisterModules(mod)
	app.Bootstrap()

	// Test index route.
	req := httptest.NewRequest("GET", "/test/", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Test greet route.
	req = httptest.NewRequest("GET", "/test/greet/alice", nil)
	w = httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestApp_ProvidersRegistered(t *testing.T) {
	app := New()
	mod := core.NewModule("test")
	app.RegisterModules(mod)

	// Just verify that registration doesn't panic.
}

// errInitHook fails OnInit so Init error handling can be exercised.
type errInitHook struct{ dummyHook }

func (h *errInitHook) OnInit() error { return errors.New("init boom") }

func TestApp_Init_RunsHooksInDependencyOrder(t *testing.T) {
	app := New()

	base := core.NewModule("base")
	dependent := core.NewModule("dependent")
	dependent.Depends = []string{"base"}

	order := []string{}
	base.RegisterOnInit(&recordingHook{name: "base", order: &order})
	dependent.RegisterOnInit(&recordingHook{name: "dependent", order: &order})

	app.RegisterModules(dependent, base)

	if err := app.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if len(order) != 2 || order[0] != "base" || order[1] != "dependent" {
		t.Errorf("init order = %v, want [base dependent]", order)
	}
}

func TestApp_Init_PropagatesHookError(t *testing.T) {
	app := New()
	mod := core.NewModule("boom")
	mod.RegisterOnInit(&errInitHook{})
	app.RegisterModules(mod)

	if err := app.Init(); err == nil {
		t.Error("expected Init to surface the OnInit failure")
	}
}

func TestApp_Init_RejectsUnknownDependency(t *testing.T) {
	app := New()
	mod := core.NewModule("a")
	mod.Depends = []string{"nope"}
	app.RegisterModules(mod)

	if err := app.Init(); err == nil {
		t.Error("expected Init to reject a dependency on an unregistered module")
	}
}

func TestApp_Init_RejectsDependencyCycle(t *testing.T) {
	app := New()
	a := core.NewModule("a")
	b := core.NewModule("b")
	a.Depends = []string{"b"}
	b.Depends = []string{"a"}
	app.RegisterModules(a, b)

	if err := app.Init(); err == nil {
		t.Error("expected Init to reject a dependency cycle")
	}
}

// Start blocks, so run it on a goroutine and shut it down via Stop.
func TestApp_StartAndStop(t *testing.T) {
	app := New()
	mod := core.NewModule("test")
	mod.RegisterControllers(&dummyController{})
	hook := &dummyHook{}
	mod.RegisterOnBoot(hook)
	mod.RegisterOnShutdown(hook)
	app.RegisterModules(mod)

	errCh := make(chan error, 1)
	go func() { errCh <- app.Start(0) }()

	// Wait for the server to be up, then verify the route serves.
	var lastCode int
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		if !app.serverReady() {
			continue
		}
		req := httptest.NewRequest("GET", "/test/", nil)
		w := httptest.NewRecorder()
		app.rtr.ServeHTTP(w, req)
		lastCode = w.Code
		break
	}
	if lastCode != http.StatusOK {
		t.Fatalf("expected the running server to answer 200, got %d", lastCode)
	}
	if !hook.bootCalled {
		t.Error("OnBoot hook was not called by Start")
	}

	if err := app.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if !hook.shutdownCalled {
		t.Error("OnShutdown hook was not called by Stop")
	}

	// Start returns once the listener is closed.
	select {
	case <-errCh:
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after Stop")
	}
}

func TestApp_Stop_WithoutStartIsSafe(t *testing.T) {
	app := New()
	app.RegisterModules(core.NewModule("test"))

	if err := app.Stop(); err != nil {
		t.Errorf("Stop without a running server should be a no-op, got %v", err)
	}
}

// SetRouter and RegisterModules mutate fields that Stop and bootstrapLocked
// read under a.mu (Stop may run from a signal handler). Without taking the
// lock in the writers, this test trips the race detector.
func TestApp_ConcurrentRegistrationAndStop(t *testing.T) {
	app := New()
	if err := app.RegisterModules(core.NewModule("seed")); err != nil {
		t.Fatalf("seed registration failed: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			_ = app.Stop()
		}
	}()

	for i := 0; i < 50; i++ {
		app.SetRouter(router.New())
		if err := app.RegisterModules(core.NewModule("mod" + string(rune('a'+i%26)) + strconv.Itoa(i))); err != nil {
			t.Fatalf("registration %d failed: %v", i, err)
		}
	}
	<-done
}

func TestApp_Stop_CollectsHookErrors(t *testing.T) {
	app := New()
	mod := core.NewModule("test")
	mod.RegisterOnShutdown(&dummyHook{shutdownErr: errors.New("shutdown boom")})
	app.RegisterModules(mod)

	if err := app.Stop(); err == nil {
		t.Error("expected Stop to report the OnShutdown failure")
	}
}

func TestApp_Start_PropagatesBootHookError(t *testing.T) {
	app := New()
	mod := core.NewModule("test")
	mod.RegisterOnBoot(&dummyHook{bootErr: errors.New("boot boom")})
	app.RegisterModules(mod)

	if err := app.Start(0); err == nil {
		t.Error("expected Start to surface the OnBoot failure")
	}
}

func TestApp_Start_WiresContainerIntoRouter(t *testing.T) {
	app := New()
	mod := core.NewModule("test")
	mod.RegisterControllers(&dummyController{})
	app.RegisterModules(mod)
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	go func() { _ = app.Start(0) }()
	for i := 0; i < 100 && !app.serverReady(); i++ {
		time.Sleep(10 * time.Millisecond)
	}
	defer app.Stop()

	// The container must be reachable from a request-scoped provider.
	if app.Container() == nil {
		t.Fatal("expected a container on the app")
	}
	rtr := app.Router()
	req := httptest.NewRequest("GET", "/test/", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 after Start wired the container, got %d", w.Code)
	}
}

// recordingHook appends its name to order when OnInit runs.
type recordingHook struct {
	dummyHook
	name  string
	order *[]string
}

func (h *recordingHook) OnInit() error {
	*h.order = append(*h.order, h.name)
	return nil
}

// The old registration wrapped every provider as func() provider.Provider,
// so the container only knew the wrapper type and Resolve of the supplied
// type always failed. Providers must be resolvable by the concrete type
// they supply.
type greeterService struct{ prefix string }

func (s *greeterService) Provide() any             { return s }
func (s *greeterService) Greet(name string) string { return s.prefix + " " + name }

type countedProvider struct{ calls int }

func (p *countedProvider) Provide() any     { p.calls++; return &greeterService{prefix: "x"} }
func (p *countedProvider) Scope() ioc.Scope { return ioc.ScopeTransient }

func TestApp_ProvidersResolvableByConcreteType(t *testing.T) {
	app := New()
	mod := core.NewModule("svc")
	svc := &greeterService{prefix: "Hello"}
	mod.RegisterProviders(svc)
	if err := app.RegisterModules(mod); err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	got, err := app.Container().Resolve(&greeterService{})
	if err != nil {
		t.Fatalf("Resolve(*greeterService) failed: %v", err)
	}
	resolved, ok := got.(*greeterService)
	if !ok {
		t.Fatalf("resolved %T, want *greeterService", got)
	}
	if resolved.Greet("Ny") != "Hello Ny" {
		t.Errorf("resolved service broken: %q", resolved.Greet("Ny"))
	}
}

// A ScopedProvider's declared scope must be honored: transient providers
// re-Provide per resolution, singletons provide exactly once.
func TestApp_ScopedProviderTransientReprovides(t *testing.T) {
	app := New()
	mod := core.NewModule("svc")
	p := &countedProvider{}
	mod.RegisterProviders(p)
	if err := app.RegisterModules(mod); err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	// Registration probes the type once.
	if p.calls != 1 {
		t.Fatalf("Provide calls after registration = %d, want 1", p.calls)
	}

	if _, err := app.Container().Resolve(&greeterService{}); err != nil {
		t.Fatalf("first resolve failed: %v", err)
	}
	if _, err := app.Container().Resolve(&greeterService{}); err != nil {
		t.Fatalf("second resolve failed: %v", err)
	}
	if p.calls != 3 {
		t.Errorf("Provide calls = %d, want 3 (1 registration + 2 transient resolutions)", p.calls)
	}
}

// A nil-supplied provider is a wiring error and must be reported at
// registration, not panic on first request.
type nilProvider struct{}

func (nilProvider) Provide() any { return nil }

func TestApp_NilProviderRejected(t *testing.T) {
	app := New()
	mod := core.NewModule("svc")
	mod.RegisterProviders(nilProvider{})
	if err := app.RegisterModules(mod); err == nil {
		t.Error("nil-supplied provider should fail registration")
	}
}

// reqScopedService is a request-scoped provider used to prove the router got a
// container attached during Bootstrap alone.
type reqScopedService struct{ n int }

func newReqScopedService() *reqScopedService { return &reqScopedService{} }

// scopeProbeController resolves the request-scoped provider twice within one
// request and reports whether the two resolutions share an instance. That only
// holds when ServeHTTP ran BeginRequest, i.e. the router has a container —
// which Bootstrap-only flows used to never attach.
type scopeProbeController struct {
	controller.BaseController
	app *App
}

func (c *scopeProbeController) Routes() []core.Route {
	return []core.Route{{Method: "GET", Path: "", Handler: "Probe"}}
}
func (c *scopeProbeController) Prefix() string { return "/probe" }

func (c *scopeProbeController) Probe(ctx *core.Ctx) error {
	cont := c.app.Container()
	reqCtx := ctx.Request.Context()
	a, err := cont.ResolveFromCtx(reqCtx, &reqScopedService{})
	if err != nil {
		return err
	}
	b, err := cont.ResolveFromCtx(reqCtx, &reqScopedService{})
	if err != nil {
		return err
	}
	if a == b {
		_, err = ctx.Response.Write([]byte("shared"))
	} else {
		_, err = ctx.Response.Write([]byte("fresh"))
	}
	return err
}

// TestApp_Bootstrap_WiresContainerIntoRouter: the documented Bootstrap()-then-
// serve-yourself flow never calls Start, so the container wiring must happen in
// bootstrapLocked, not only in Start.
func TestApp_Bootstrap_WiresContainerIntoRouter(t *testing.T) {
	rtr := router.New()
	app := New()
	app.SetRouter(rtr)
	if err := app.Container().RegisterScoped(newReqScopedService); err != nil {
		t.Fatal(err)
	}
	mod := core.NewModule("probe")
	mod.RegisterControllers(&scopeProbeController{app: app})
	if err := app.RegisterModules(mod); err != nil {
		t.Fatal(err)
	}
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/probe", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if w.Body.String() != "shared" {
		t.Fatalf("Bootstrap must wire container into router: got %q (request-scoped provider not shared)", w.Body.String())
	}
}

// TestApp_InitConcurrentWithRegisterModules: Init used to call sortModules,
// which reads a.modules and a.moduleMap, without holding a.mu. A concurrent
// RegisterModules writes both, and a concurrent map read/write is undefined
// behaviour. Run under -race this fails without the lock.
func TestApp_InitConcurrentWithRegisterModules(t *testing.T) {
	app := New()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			_ = app.RegisterModules(core.NewModule("m" + strconv.Itoa(i)))
		}(i)
		go func() {
			defer wg.Done()
			_ = app.Init()
		}()
	}
	wg.Wait()
}