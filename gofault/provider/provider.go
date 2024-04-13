// Package provider provides utilities for defining injectable providers.
package provider

import "github.com/gofault/gofault/ioc"

// Provider is the interface for injectable services.
type Provider interface {
	Provide() any
}

// ProviderFunc is an adapter that turns a function into a Provider.
type ProviderFunc func() any

// Provide returns the result of calling the underlying function.
func (f ProviderFunc) Provide() any {
	return f()
}

// ValueProvider wraps a concrete value so it can be registered in the container.
type ValueProvider struct {
	Value any
}
