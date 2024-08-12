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