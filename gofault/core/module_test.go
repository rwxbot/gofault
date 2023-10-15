package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewCtx(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewCtx(w, r)
	if ctx.Request != r {
		t.Error("Request not set")
	}
	if ctx.Response != w {
		t.Error("Response not set")
	}
	if ctx.StatusCode != http.StatusOK {
		t.Errorf("expected StatusOK, got %d", ctx.StatusCode)
	}
	if ctx.Params == nil {
		t.Error("Params should be initialized")
	}
	if ctx.Locals == nil {
		t.Error("Locals should be initialized")
	}
}

func TestCtx_Locals(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)
	ctx := NewCtx(w, r)

	// Store and retrieve
	ctx.Locals["user"] = "alice"
	ctx.Locals["score"] = 42

	if v := ctx.Locals["user"]; v != "alice" {
		t.Errorf("expected alice, got %v", v)
	}
	if v := ctx.Locals["score"]; v != 42 {
		t.Errorf("expected 42, got %v", v)
	}
}

func TestNewModule(t *testing.T) {
	m := NewModule("test")
	if m.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", m.Name)
	}
	if m.Depends == nil {
		t.Error("Depends should be initialized")
	}
	if m.Controllers == nil {
		t.Error("Controllers should be initialized")
	}
	if m.Providers == nil {
		t.Error("Providers should be initialized")
	}
	if m.Middleware == nil {
		t.Error("Middleware should be initialized")
	}
}

func TestModule_RegisterControllers(t *testing.T) {
	m := NewModule("test")
	ctrl := &testController{}

	m.RegisterControllers(ctrl)
	if len(m.Controllers) != 1 {
		t.Errorf("expected 1 controller, got %d", len(m.Controllers))
	}
	m.RegisterControllers(ctrl, ctrl)
	if len(m.Controllers) != 3 {
		t.Errorf("expected 3 controllers, got %d", len(m.Controllers))
	}
}

func TestModule_RegisterProviders(t *testing.T) {
	m := NewModule("test")
	prov := &testProvider{}

	m.RegisterProviders(prov)
	if len(m.Providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(m.Providers))
	}
}

func TestModule_RegisterMiddleware(t *testing.T) {
	m := NewModule("test")