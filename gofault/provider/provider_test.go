package provider

import (
	"testing"

	"github.com/gofault/gofault/ioc"
)

func TestProviderFunc(t *testing.T) {
	val := "not-called"
	pf := ProviderFunc(func() any {
		val = "called"
		return "hello"
	})

	// Function body should not run at construction time
	if val != "not-called" {
		t.Error("ProviderFunc should not run at construction")
	}

	// Function body runs on Provide()
	pf.Provide()
	if val != "called" {
		t.Error("ProviderFunc not called on Provide()")
	}

	result := pf.Provide()
	if result != "hello" {