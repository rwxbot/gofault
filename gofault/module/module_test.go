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
