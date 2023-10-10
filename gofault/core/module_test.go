package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewCtx(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewCtx(w, r)
	if ctx.Request != r {