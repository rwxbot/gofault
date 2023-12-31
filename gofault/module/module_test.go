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