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
