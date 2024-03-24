package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

func dummyHandler(ctx *core.Ctx) error {
	return nil
}

func TestRouter_HandleAndServe(t *testing.T) {