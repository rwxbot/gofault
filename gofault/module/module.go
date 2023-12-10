// Package module provides the application root that bootstraps all registered modules.
package module

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/gofault/gofault/controller"
	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/ioc"
	"github.com/gofault/gofault/provider"
	"github.com/gofault/gofault/router"
	"github.com/gofault/gofault/server"
)

// App is the root application structure that ties everything together.
type App struct {
	container *ioc.Container
	rtr       *router.Router
	modules   []*core.Module // sorted by dependency order
	moduleMap map[string]*core.Module

	// mu guards server, booted and the router wiring below them. Start runs on
	// the caller's goroutine while Stop is conventionally invoked from a
	// signal handler, so those fields are written and read concurrently.
	mu     sync.Mutex
	server *server.HTTP
	booted bool
}

// New creates a new application instance.
func New() *App {
	return &App{