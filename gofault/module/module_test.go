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