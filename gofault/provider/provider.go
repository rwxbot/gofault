// Package provider provides utilities for defining injectable providers.
package provider

import "github.com/gofault/gofault/ioc"

// Provider is the interface for injectable services.
type Provider interface {
	Provide() any
}