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