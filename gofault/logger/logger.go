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
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// defaultLogger is a simple logger implementation.
type defaultLogger struct {
	logger *log.Logger
	level  int
}

const (
	LevelDebug = 0
	LevelInfo  = 1
	LevelWarn  = 2
	LevelError = 3
)

// New creates a new logger instance.
func New(prefix string, level int) Logger {
	return &defaultLogger{
		logger: log.New(os.Stdout, prefix, log.LstdFlags),
		level:  level,
	}
}

func (l *defaultLogger) log(level int, msg string, args ...any) {
	if level < l.level {
		return
	}
	formatted := msg