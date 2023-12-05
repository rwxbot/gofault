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