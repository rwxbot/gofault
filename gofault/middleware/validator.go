package middleware

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

// ValidationError represents a single validation error.