# gofault

A modular backend framework for Go: modules, dependency injection and an HTTP
layer that stays out of the way.

- **Modules** compose the application and declare their dependencies
- **Controllers** expose HTTP endpoints through a prefix and a route table
- **Providers** supply dependencies through the IoC container
- **Middleware** wraps the request pipeline and can short-circuit it

## Features

| Area | Package |
|------|---------|
| IoC container (singleton, request and transient scope) | `ioc` |
| Module registry with topological ordering | `module`, `core` |
| Router with `:param` matching | `router` |
| Controller prefix routing and JSON helpers | `controller` |
| Lifecycle hooks (`OnInit`, `OnBoot`, `OnShutdown`) | `core` |
| Exception filter | `exception` |
| Structured logging | `logger` |
| YAML/JSON config loading | `config` |

## Install

```bash
go get github.com/gofault/gofault
```

## Quick start

```go
package main

import (
    "log"

    "github.com/gofault/gofault/controller"
    "github.com/gofault/gofault/core"
    "github.com/gofault/gofault/module"
)

type GreetingController struct {
    controller.BaseController
    prefix string
}

func (c *GreetingController) Prefix() string { return "/hello" }

func (c *GreetingController) Routes() []core.Route {
    return []core.Route{
        {Method: "GET", Path: "/greet/:name", Handler: "Greet"},
    }
}

func (c *GreetingController) Greet(ctx *core.Ctx) error {
    return controller.OK(ctx.Response, map[string]string{
        "message": c.prefix + ", " + controller.Param(ctx, "name") + "!",
    })
}

func main() {
    mod := core.NewModule("hello")
    mod.RegisterControllers(&GreetingController{prefix: "Hello"})

    app := module.New()
    if err := app.RegisterModules(mod); err != nil {
        log.Fatal(err)
    }
    log.Fatal(app.Start(9090))
}
```

```
$ curl localhost:9090/hello/greet/World
{"code":0,"message":"success","data":{"message":"Hello, World!"}}
```

## Architecture

Dependencies point in one direction: every package imports `core`, and `core`
imports nothing.

```
core          contracts only (Ctx, Module, Route, Provider, Controller, hooks)
 ├── ioc      container: singleton / request / transient scopes
 ├── provider injectable values, including request-scoped ones
 ├── router   method + path dispatch, middleware chain, exception filter
 ├── server   net/http listener and graceful shutdown
 └── module   composition root: orders modules, wires routes, starts the server
      ├── middleware   cross-cutting handlers (logging, recovery, timeout, ...)
      └── controller   prefix routing plus response helpers
```

`module.App` is the composition root. It topologically sorts the registered
modules by their `Depends` list, rejects unknown or circular dependencies, then
wires every controller route onto the router with that module's middleware:

```go
app := module.New()
app.SetRouter(rtr)

if err := app.RegisterModules(modA, modB); err != nil {
    log.Fatal(err)
}
log.Fatal(app.Start(9090))
```

## Testing

```bash
go test ./... -race
```

Integration tests that need an external service skip themselves when it is not
reachable, so the suite stays green on a bare checkout.

## What's new in v0.1.0

The first cut of the framework: the `core` contracts, the IoC container and the
module primitives.

### Contracts

```go
type Handler func(ctx *Ctx) error
type MiddlewareFunc func(ctx *Ctx, next Handler) error

type Provider interface{ Provide() any }

type Controller interface {
    Routes() []Route
    Prefix() string
}
```

`Ctx` wraps the request/response pair plus the params extracted from the route:

```go
type Ctx struct {
    Request    *http.Request
    Response   http.ResponseWriter
    Params     map[string]string
    StatusCode int
    Locals     map[string]any
}
```

### Modules

```go
mod := core.NewModule("hello")
mod.RegisterControllers(ctrl).RegisterProviders(svc)
```

### Container

```go
c := ioc.New()
c.Register(NewGreetingService)

svc, err := c.Resolve(&GreetingService{})
```

## What's new in v0.5.0

Middleware chains, lifecycle hooks and a filter for error handling.

### Middleware

A middleware wraps the next handler; returning an error stops the chain and
hands the error to the router's exception filter:

```go
func loggingMiddleware(ctx *core.Ctx, next core.Handler) error {
    log.Printf("%s %s", ctx.Request.Method, ctx.Request.URL.Path)
    return next(ctx)
}
```

### Lifecycle hooks

Three small interfaces let a type participate in startup and shutdown. They are
optional: implement whichever you need.

```go
type OnInit interface{ OnInit() error }
type OnBoot interface{ OnBoot() error }
type OnShutdown interface{ OnShutdown() error }
```

Hooks are registered on the module and run in dependency order, with shutdown
running in reverse:

```go
mod.RegisterOnBoot(hook)
mod.RegisterOnShutdown(hook)
```

### Exceptions

`exception` carries the HTTP status alongside the message, so handlers can
return an error and let the filter turn it into a response:

```go
return exception.BadRequest("invalid payload")
```


## What's new in v1.0.0

### Application bootstrap

`module.App` is the composition root. It sorts modules by their `Depends`
declarations, rejects unknown and circular dependencies, wires every controller
route onto the router, and then starts the server:

```go
app := module.New()
app.SetRouter(rtr)

if err := app.RegisterModules(mod); err != nil {
    log.Fatal(err)
}
log.Fatal(app.Start(9090))
```

### Server, router and controller

`server.HTTP` owns the listener, `router` dispatches on method and path, and
`controller` provides the response helpers:

```go
rtr := router.New()
rtr.Middleware(loggingMiddleware)
rtr.ExceptionFilter(exception.NewHTTPExceptionFilter())

return controller.OK(ctx.Response, map[string]string{"message": "gofault is running"})
```

### Module dependencies

`Depends` orders initialisation without constraining registration order:

```go
mod.Depends = []string{"config"}
```

### Injectable providers

```go
mod.RegisterProviders(provider.ProviderFunc(func() any { return svc }))
```


## What's new in v1.1.0

### Configuration

`config.Load` reads a YAML document into typed sections; `MustLoad` panics
instead of returning an error:

```go
cfg, err := config.Load("config.yaml")
if err != nil {
    return err
}
addr := cfg.Server.GetAddress()
dsn := cfg.Database.GetDSN()
```

### HTTP exception filter

`exception.NewHTTPExceptionFilter` maps an `exception.HTTPException` back to its
status code. Wire it onto the router so handler errors become responses:

```go
rtr.ExceptionFilter(exception.NewHTTPExceptionFilter())
```

Filters compose — build a chain when you need more than one:

```go
chain := exception.NewExceptionFilterChain(
    exception.NewHTTPExceptionFilter(),
)
```


## What's new in v1.2.0

### Graceful shutdown

`App.Stop` runs every `OnShutdown` hook in reverse module order and then drains
the server, so in-flight requests are not cut off:

```go
defer app.Stop()
```

`server.HTTP` owns the listener and exposes the same behaviour directly:

```go
srv := server.New(rtr, 9090)
go srv.Start()

server.RegisterShutdownHook(func() error { return srv.Stop() })
```

### Example application

`examples/hello` is a runnable application that exercises a provider, a
controller with a path parameter, module middleware and both lifecycle hooks.


## What's new in v1.3.0

### Route grouping and parameter binding

A controller owns a prefix and a route table; `:name` segments become request
params:

```go
func (c *AdminController) Prefix() string { return "/admin" }

func (c *AdminController) Routes() []core.Route {
    return []core.Route{
        {Method: "GET", Path: "/users", Handler: "ListUsers"},
        {Method: "POST", Path: "/users/:id", Handler: "UpdateUser"},
    }
}
```

Handlers read them through the `controller` helpers:

```go
id := controller.Param(ctx, "id")
filter := controller.Query(ctx, "filter")
```

### Continuous integration

`.github/workflows/ci.yml` runs the test suite with the race detector, lints
with `golangci-lint`, and publishes release binaries for Linux, macOS and
Windows when a `v*` tag is pushed. The workflow pins the Go version that
`go.mod` declares.


## What's new in v1.4.0

### Unified response encoding

Every JSON response goes through the same envelope, and every helper sets
`Content-Type: application/json` before the status line:

```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

```go
return controller.OK(ctx.Response, user)                      // 200
return controller.Error(ctx.Response, 422, "invalid payload") // 422
```

`BaseController` exposes the same helpers as methods, plus a plain-text variant
for non-JSON endpoints:

```go
c.JSON(ctx.Response, http.StatusOK, user)
c.Text(ctx.Response, http.StatusOK, "pong")
```


## What's new in v1.5.0

### Request validation

`ValidateRequest` checks a set of named rules and returns an error listing every
failure:

```go
err := middleware.ValidateRequest(ctx, middleware.Rules{
    "email": {middleware.Required(), middleware.Email()},
    "age":   {middleware.Min(18), middleware.Max(120)},
})
if err != nil {
    return exception.BadRequest(err.Error())
}
```

Available rules: `Required`, `Min`, `Max`, `MinLength`, `MaxLength`, `Regex`,
`Email`.

### Health checks

`HealthCheckMiddleware` answers a probe path and reports each registered check.
`LatencyHealthCheck` wraps any function and records how long it took:

```go
cfg := middleware.DefaultHealthCheckConfig()
cfg.Path = "/healthz"