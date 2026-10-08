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

// Provide returns the stored value.
func (v ValueProvider) Provide() any {
	return v.Value
}

// ScopedProvider is implemented by providers that declare their own lifecycle scope.
// If a provider implements this interface, its declared scope takes precedence
// over the registration method used (Register/RegisterScoped/RegisterTransient).
type ScopedProvider interface {
	Provider
	// Scope returns the lifecycle scope for this provider.
	Scope() ioc.Scope
}
