// Package ioc provides the inversion-of-control container for dependency injection.
package ioc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// ErrDuplicateRegistration is returned when two constructors are registered for
// the same type in the same scope. Silently replacing the first binding made a
// later typo look like it had taken effect.
var ErrDuplicateRegistration = errors.New("duplicate registration")

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