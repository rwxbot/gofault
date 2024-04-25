// Package logger provides logging middleware for gofault.
package logger

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofault/gofault/core"
)

// Logger defines the logging interface.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)