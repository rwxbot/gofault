// Package ioc provides the inversion-of-control container for dependency injection.
package ioc

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// contextKey is a custom type to avoid collisions in context.WithValue.
type contextKey string

const requestCtxKey contextKey = "gofault-request-scope"

// Scope defines the lifecycle scope of a registered service.
type Scope int

const (
	ScopeSingleton Scope = iota
	ScopeRequest
	ScopeTransient
)

// Container is the publicly exposed IoC container type (alias of the internal container).
type Container = container

// container is the inversion-of-control container.
type container struct {
	mu         sync.RWMutex
	singletons map[reflect.Type]*entry
	transients map[reflect.Type]*entry
	request    map[reflect.Type]*entry
	// requestScope stores per-context instances for request-scoped resolution.
	// The key is the context value returned by context.WithValue.
	requestScope map[contextKey]map[reflect.Type]any
	// counter for generating unique request scope keys.
	scopeCounter uint64
}

type entry struct {
	ctor any
	inst any // nil means not-yet-instantiated (for singletons)