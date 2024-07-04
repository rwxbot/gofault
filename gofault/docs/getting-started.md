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