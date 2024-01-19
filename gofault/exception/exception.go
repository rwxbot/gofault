// Package exception provides HTTP exception types and exception filtering.
package exception

import (
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