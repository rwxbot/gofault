// Package module provides the application root that bootstraps all registered modules.
package module

import (
	"fmt"

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
	server    *server.HTTP
	booted    bool
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
func (a *App) SetRouter(r *router.Router) {
	a.rtr = r
}

// RegisterModules registers one or more application modules.
func (a *App) RegisterModules(mods ...*core.Module) error {
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
// avoiding the classic Go closure capture bug where loop variables are
// captured by reference instead of by value.
func makeHandler(ctrl core.Controller, route core.Route) core.Handler {
	return func(ctx *core.Ctx) error {
		return controller.InvokeHandler(ctrl, route.Method, route.Path, route.Handler, ctx)
	}
}

// sortModules topological sorts modules by their Depends declarations.
// Modules with no dependencies or only resolved dependencies come first.
func (a *App) sortModules() error {
	// Build dependency graph and check for missing dependencies.
	for _, m := range a.modules {
		for _, dep := range m.Depends {
			if _, exists := a.moduleMap[dep]; !exists {
				return fmt.Errorf("module %q depends on unknown module %q", m.Name, dep)
			}
		}
	}

	// Kahn's algorithm for topological sort.
	var sorted []*core.Module
	resolved := make(map[string]bool)
	remaining := make(map[string]*core.Module)
	for _, m := range a.modules {
		remaining[m.Name] = m
	}

	for len(remaining) > 0 {
		progress := false
		for name, m := range remaining {
			allResolved := true
			for _, dep := range m.Depends {
				if !resolved[dep] {
					allResolved = false