//go:build soak

package soak

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofault/gofault/controller"
	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

// echoController mirrors the e2e controller: the common path, a large body, a
// parameterised route and a deliberate failure.
type echoController struct {
	controller.BaseController
	payloadSize int
}

func (c *echoController) Routes() []core.Route {
	return []core.Route{
		{Method: http.MethodGet, Path: "/echo", Handler: "Echo"},
		{Method: http.MethodGet, Path: "/big", Handler: "Big"},
		{Method: http.MethodGet, Path: "/items/:id", Handler: "Item"},
		{Method: http.MethodGet, Path: "/fail", Handler: "Fail"},
	}
}

func (c *echoController) Prefix() string { return "/" }

func (c *echoController) Echo(ctx *core.Ctx) error {
	return controller.OK(ctx.Response, map[string]string{"path": ctx.Request.URL.Path})
}

func (c *echoController) Big(ctx *core.Ctx) error {
	ctx.Response.Header().Set("Content-Type", "text/plain")
	_, err := ctx.Response.Write([]byte(strings.Repeat("x", c.payloadSize)))
	return err
}

func (c *echoController) Item(ctx *core.Ctx) error {
	return controller.OK(ctx.Response, map[string]string{"id": ctx.Params["id"]})
}

func (c *echoController) Fail(ctx *core.Ctx) error {
	return exception.BadRequest("deliberate failure")
}

// slowController is the one that holds a goroutine per in-flight request, so a
// leak in the timeout or handler path would show as a goroutine slope.
type slowController struct {
	delay time.Duration
}

func (c *slowController) Routes() []core.Route {
	return []core.Route{
		{Method: http.MethodGet, Path: "/slow", Handler: "Slow"},
	}
}

func (c *slowController) Prefix() string { return "/" }

func (c *slowController) Slow(ctx *core.Ctx) error {
	select {
	case <-time.After(c.delay):
	case <-ctx.Request.Context().Done():
		return ctx.Request.Context().Err()
	}
	return controller.OK(ctx.Response, "slow done")
}

func wrap(c core.Controller, method string) core.Handler {
	return func(ctx *core.Ctx) error {
		return controller.InvokeHandler(c, method, ctx)
	}
}
