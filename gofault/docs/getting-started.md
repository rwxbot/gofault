# Getting started

This guide takes you from an empty module to a running service, then explains how
to test it.

## Install

```bash
go get github.com/gofault/gofault
```

The framework targets the Go version declared in `go.mod`.

## A minimal application

An application is a set of modules. A module carries controllers, providers,
middleware and lifecycle hooks; a controller carries a route table.

```go
package main

import (
    "log"

    "github.com/gofault/gofault/controller"
    "github.com/gofault/gofault/core"
    "github.com/gofault/gofault/module"
)

// GreetingService is an injectable provider.
type GreetingService struct {
    prefix string
}

func (s *GreetingService) Provide() any { return s }

func (s *GreetingService) Greet(name string) string {
    return s.prefix + ", " + name + "!"
}

type GreetingController struct {
    controller.BaseController
    svc *GreetingService
}

func NewGreetingController(svc *GreetingService) *GreetingController {
    return &GreetingController{svc: svc}
}

func (c *GreetingController) Prefix() string { return "/hello" }

func (c *GreetingController) Routes() []core.Route {
    return []core.Route{
        {Method: "GET", Path: "/", Handler: "Index"},
        {Method: "GET", Path: "/greet/:name", Handler: "Greet"},
    }
}

func (c *GreetingController) Index(ctx *core.Ctx) error {
    return controller.OK(ctx.Response, map[string]string{"message": "gofault is running"})
}

func (c *GreetingController) Greet(ctx *core.Ctx) error {
    return controller.OK(ctx.Response, map[string]string{
        "message": c.svc.Greet(controller.Param(ctx, "name")),
    })
}

func main() {
    svc := &GreetingService{prefix: "Hello"}
    ctrl := NewGreetingController(svc)

    mod := core.NewModule("hello")
    mod.RegisterControllers(ctrl)
    mod.RegisterProviders(svc)

    app := module.New()
    if err := app.RegisterModules(mod); err != nil {
        log.Fatal(err)
    }
    log.Fatal(app.Start(9090))
}
```

Run it and call the routes:

```bash
go run .
curl localhost:9090/hello/
# {"code":0,"message":"success","data":{"message":"gofault is running"}}
curl localhost:9090/hello/greet/World
# {"code":0,"message":"success","data":{"message":"Hello, World!"}}
```

## Adding middleware and error handling

Middleware lives on the module, so it only applies to that module's routes. The
exception filter lives on the router, because it has to catch errors from every
route:

```go
import (
    "github.com/gofault/gofault/exception"
    "github.com/gofault/gofault/middleware"
    "github.com/gofault/gofault/router"
)

func loggingMiddleware(ctx *core.Ctx, next core.Handler) error {
    log.Printf("%s %s", ctx.Request.Method, ctx.Request.URL.Path)
    return next(ctx)
}

mod.RegisterMiddleware(
    middleware.RecoveryMiddleware(middleware.DefaultRecoveryConfig()),
    loggingMiddleware,
)

rtr := router.New()
rtr.ExceptionFilter(exception.NewHTTPExceptionFilter())
app.SetRouter(rtr)
```
