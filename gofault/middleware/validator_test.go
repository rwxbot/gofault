package middleware

import (
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

func TestRequired(t *testing.T) {
	r := &Required{}

	// Valid case
	err := r.Validate("field", "value")
	if err != nil {
		t.Errorf("expected no error for non-empty value, got %v", err)
	}

	// Empty string
	err = r.Validate("field", "")
	if err == nil {
		t.Error("expected error for empty string")
	}

	// Nil
	err = r.Validate("field", nil)
	if err == nil {
		t.Error("expected error for nil")
	}
}
