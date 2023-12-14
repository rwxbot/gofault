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
		container: ioc.New(),
		rtr:       router.New(),
		modules:   make([]*core.Module, 0),
		moduleMap: make(map[string]*core.Module),
	}
}

// SetRouter replaces the default router with a custom one (e.g., to add filters).
// The write takes a.mu: Stop and bootstrapLocked read a.rtr under the same lock,
// and Stop may run concurrently from a signal handler.
func (a *App) SetRouter(r *router.Router) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.rtr = r
}

// RegisterModules registers one or more application modules.
// The mutation takes a.mu: a.modules and a.moduleMap are read under the same
// lock by bootstrapLocked and by Stop (which may run from a signal handler).
func (a *App) RegisterModules(mods ...*core.Module) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, m := range mods {
		if m.Name == "" {
			return fmt.Errorf("module name cannot be empty")
		}
		if _, exists := a.moduleMap[m.Name]; exists {
			return fmt.Errorf("duplicate module name: %q", m.Name)
		}
		a.moduleMap[m.Name] = m
		for _, ctrl := range m.Controllers {
			a.container.Register(func() core.Controller { return ctrl })
		}
		for _, prov := range m.Providers {
			a.container.Register(func() provider.Provider { return prov })
		}
		a.modules = append(a.modules, m)
	}
	return nil
}

// makeHandler creates a handler closure that properly captures route values,