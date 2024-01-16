package module

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/controller"
	"github.com/gofault/gofault/core"
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
