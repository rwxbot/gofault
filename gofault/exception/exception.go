// Package exception provides HTTP exception types and exception filtering.
package exception

import (
	"errors"
	"fmt"
	"net/http"
)

// HTTPException is the base error type for HTTP-related exceptions.
type HTTPException interface {
	error
	GetStatusCode() int
	GetMessage() string
	GetCode() int
}

// baseException implements HTTPException.
type baseException struct {
	statusCode int
	code       int
	message    string
}

func (e *baseException) Error() string {
	return e.message
}

func (e *baseException) GetStatusCode() int { return e.statusCode }
func (e *baseException) GetMessage() string { return e.message }
func (e *baseException) GetCode() int       { return e.code }

// ExceptionFilter is the interface for exception handling.
type ExceptionFilter interface {
	// ExceptionCapture is called when an exception occurs.
	// Return true if the exception was handled (and should not propagate).
	// Return false to let the exception propagate to the next filter.
	Capture(ctx any, err error) bool
}

// ExceptionHandler is the function-based form of ExceptionFilter.
type ExceptionHandler func(ctx any, err error) bool

func (h ExceptionHandler) Capture(ctx any, err error) bool {
	return h(ctx, err)
}

// ExceptionFilterChain holds multiple filters and executes them in order.
type ExceptionFilterChain struct {
	filters []ExceptionFilter
}

func NewExceptionFilterChain(filters ...ExceptionFilter) *ExceptionFilterChain {
	return &ExceptionFilterChain{filters: filters}
}

func (c *ExceptionFilterChain) Add(filters ...ExceptionFilter) {
	c.filters = append(c.filters, filters...)
}

// Capture runs all filters until one handles the exception.
func (c *ExceptionFilterChain) Capture(ctx any, err error) bool {
	for _, f := range c.filters {
		if f.Capture(ctx, err) {
			return true
		}
	}
	return false
}

// CaptureException runs the chain and returns the exception if unhandled.
func (c *ExceptionFilterChain) CaptureException(ctx any, err error) error {
	handled := c.Capture(ctx, err)
	if !handled {
		return err
	}
	return nil
}

// BadRequest creates a 400 Bad Request exception.
func BadRequest(message string) error {
	return &baseException{statusCode: http.StatusBadRequest, code: 400, message: message}
}

// Unauthorized creates a 401 Unauthorized exception.
func Unauthorized(message string) error {
	return &baseException{statusCode: http.StatusUnauthorized, code: 401, message: message}
}

// Forbidden creates a 403 Forbidden exception.
func Forbidden(message string) error {
	return &baseException{statusCode: http.StatusForbidden, code: 403, message: message}
}

// NotFound creates a 404 Not Found exception.
func NotFound(message string) error {
	return &baseException{statusCode: http.StatusNotFound, code: 404, message: message}
}

// InternalServerError creates a 500 Internal Server Error exception.
func InternalServerError(message string) error {
	return &baseException{statusCode: http.StatusInternalServerError, code: 500, message: message}
}

// Conflict creates a 409 Conflict exception.
func Conflict(message string) error {
	return &baseException{statusCode: http.StatusConflict, code: 409, message: message}
}

// MethodNotAllowed creates a 405 Method Not Allowed exception.