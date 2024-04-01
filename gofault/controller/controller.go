// Package controller provides the controller base type and HTTP method helpers.
package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gofault/gofault/core"
)

// BaseController provides helper methods for writing HTTP handlers.
type BaseController struct{}

// JSON writes a JSON response with the given status code.
func (c *BaseController) JSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// Text writes a plain text response.
func (c *BaseController) Text(w http.ResponseWriter, status int, msg string) error {
	w.WriteHeader(status)
	_, err := w.Write([]byte(msg))