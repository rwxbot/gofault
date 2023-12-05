# gofault

A modular backend framework for Go, built around dependency injection and
explicit composition.

> Status: prototype. The container and the module primitives are in place; the
> HTTP layer is still landing.

## Install

```bash
go get github.com/gofault/gofault
```

## Core interfaces

`core` is the contract package. It imports nothing else in the framework, so
every other package can depend on it without creating a cycle:

```go
type Provider interface{ Provide() any }

type Controller interface {
    Routes() []Route
    Prefix() string
}

type Route struct {
    Method  string
    Path    string
    Handler string // name of the controller method, e.g. "Greet"
}

type Handler func(ctx *Ctx) error
type MiddlewareFunc func(ctx *Ctx, next Handler) error
```

A module is a plain struct rather than an interface to implement:

```go
mod := core.NewModule("hello")
mod.RegisterControllers(ctrl)
mod.RegisterProviders(svc)
```

## The container

`ioc` resolves constructors by type, which keeps wiring explicit and testable.
`Resolve` takes a pointer to a zero value of the type you want:

```go
c := ioc.New()
c.Register(NewGreetingService)

svc, err := c.Resolve(&GreetingService{})
```


## What's new in v0.1.0

The first cut of the framework: the `core` contracts, the IoC container and the
module primitives.

### Contracts