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

// Router matches incoming requests against registered routes and executes the middleware chain.
type Router struct {
	middleware      []core.MiddlewareFunc
	routes          []routeEntry
	exceptionFilter exception.ExceptionFilter
	container       *ioc.Container
}

type routeEntry struct {
	method     string
	pattern    *regexp.Regexp
	paramNames []string
	handler    core.Handler
	middleware []core.MiddlewareFunc
}

var _ RouterInterface = (*Router)(nil)

// RouterInterface defines the routing operations exposed to the server.
type RouterInterface interface {
	Handle(method, path string, handler core.Handler, mw ...core.MiddlewareFunc)
	Middleware(mw ...core.MiddlewareFunc)
	ExceptionFilter(f exception.ExceptionFilter)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}