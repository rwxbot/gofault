// Package router provides HTTP routing with middleware chain support.
package router

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
	"github.com/gofault/gofault/ioc"
)
