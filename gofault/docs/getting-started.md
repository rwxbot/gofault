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