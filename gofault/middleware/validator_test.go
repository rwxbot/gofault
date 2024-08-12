package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

func TestRequired(t *testing.T) {
	r := &Required{}

	// Valid case
	err := r.Validate("field", "value")
	if err != nil {