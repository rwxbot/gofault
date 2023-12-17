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
			if err := a.registerProvider(prov); err != nil {
				return fmt.Errorf("module %q: %w", m.Name, err)
			}
		}
		a.modules = append(a.modules, m)
	}
	return nil
}

// registerProvider makes a provider's supplied value resolvable by its
// concrete type. The old registration — Register(func() provider.Provider
// { return prov }) — only ever exposed the wrapper type, so the container
// never saw what Provide() actually supplies and Resolve(&Service{}) always
// failed: the whole Provider mechanism was dead code.
//
// Provide() must run once at registration to learn the concrete type, which
// is the container's only key. Singletons reuse that first value (Provide
// runs exactly once); request/transient providers re-Provide per resolution
// (1+N calls). A ScopedProvider can pick its lifecycle scope.
func (a *App) registerProvider(prov provider.Provider) error {
	value := prov.Provide()
	if value == nil {
		return fmt.Errorf("provider %T supplied a nil value", prov)
	}
	t := reflect.TypeOf(value)

	scope := ioc.ScopeSingleton
	if sp, ok := prov.(provider.ScopedProvider); ok {
		scope = sp.Scope()
	}

	ctor := reflect.MakeFunc(
		reflect.FuncOf(nil, []reflect.Type{t}, false),
		func([]reflect.Value) []reflect.Value {
			v := value
			if scope != ioc.ScopeSingleton {
				v = prov.Provide()
			}
			vv := reflect.ValueOf(v)
			if !vv.IsValid() || vv.Type() != t {
				panic(fmt.Sprintf("provider %T changed supplied type: %v -> %v", prov, t, vv.Type()))
			}
			return []reflect.Value{vv}
		},
	).Interface()

	switch scope {
	case ioc.ScopeSingleton:
		return a.container.Register(ctor)
	case ioc.ScopeRequest:
		return a.container.RegisterScoped(ctor)
	case ioc.ScopeTransient:
		return a.container.RegisterTransient(ctor)
	default:
		return fmt.Errorf("provider %T declared unknown scope %d", prov, scope)
	}
}

// makeHandler creates a handler closure that properly captures route values,
// avoiding the classic Go closure capture bug where loop variables are
// captured by reference instead of by value.
func makeHandler(ctrl core.Controller, route core.Route) core.Handler {
	return func(ctx *core.Ctx) error {
		return controller.InvokeHandler(ctrl, route.Handler, ctx)
	}
}

// sortModules topological sorts modules by their Depends declarations.
//
// The result is deterministic: modules keep their registration order among
// themselves, so the same dependency graph always produces the same start
// order. Iterating a map here leaked Go's randomised iteration into observable
// behaviour, so modules booted in a different order on every run.
func (a *App) sortModules() error {
	// Build dependency graph and check for missing dependencies.
	for _, m := range a.modules {
		for _, dep := range m.Depends {
			if _, exists := a.moduleMap[dep]; !exists {
				return fmt.Errorf("module %q depends on unknown module %q", m.Name, dep)
			}
		}
	}

	// Kahn's algorithm for topological sort, walking the registration-ordered
	// slice rather than a map so the output is stable.
	var sorted []*core.Module
	resolved := make(map[string]bool, len(a.modules))
	remaining := make(map[string]*core.Module, len(a.modules))
	for _, m := range a.modules {
		remaining[m.Name] = m
	}

	for len(remaining) > 0 {
		progress := false
		for _, m := range a.modules {
			if resolved[m.Name] {
				continue
			}
			allResolved := true
			for _, dep := range m.Depends {
				if !resolved[dep] {
					allResolved = false
					break
				}
			}
			if allResolved {
				sorted = append(sorted, m)
				resolved[name] = true
				delete(remaining, name)
				progress = true