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
		t.Errorf("expected 'hello', got '%v'", result)
	}
}

func TestProviderFunc_ReturnsNil(t *testing.T) {
	pf := ProviderFunc(func() any {
		return nil
	})

	val := pf.Provide()
	if val != nil {
		t.Errorf("expected nil, got '%v'", val)
	}
}

func TestValueProvider(t *testing.T) {
	pv := ValueProvider{Value: 42}

	val := pv.Provide()
	if val != 42 {
		t.Errorf("expected 42, got %v", val)
	}
}

func TestValueProvider_Struct(t *testing.T) {
	pv := ValueProvider{Value: "custom-value"}

	if pv.Value != "custom-value" {
		t.Errorf("expected 'custom-value', got '%v'", pv.Value)
	}
}

// testScopedProvider implements both Provider and ScopedProvider for testing
type testScopedProvider struct {
	scope ioc.Scope
}
