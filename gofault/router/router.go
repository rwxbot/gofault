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

// New creates a new Router.
func New() *Router {
	return &Router{routes: make([]routeEntry, 0)}
}

// SetContainer binds an IoC container to the router for request scope management.
func (r *Router) SetContainer(c *ioc.Container) {
	r.container = c
}

// Middleware appends global middleware to the router.
func (r *Router) Middleware(mw ...core.MiddlewareFunc) {
	r.middleware = append(r.middleware, mw...)
}

// ExceptionFilter sets the exception filter for the router.
func (r *Router) ExceptionFilter(f exception.ExceptionFilter) {
	r.exceptionFilter = f
}

// Handle registers a route with the given method, path pattern, handler, and optional middleware.
func (r *Router) Handle(method, path string, handler core.Handler, mw ...core.MiddlewareFunc) {
	pattern, names := buildPattern(path)
	entry := routeEntry{
		method:     method,
		pattern:    pattern,
		paramNames: names,
		handler:    handler,
		middleware: mw,
	}
	r.routes = append(r.routes, entry)
}

func buildPattern(path string) (*regexp.Regexp, []string) {
	parts := strings.Split(path, "/")
	paramNames := []string{}
	patternParts := []string{""}

	for _, part := range parts[1:] {
		if strings.HasPrefix(part, ":") {
			paramNames = append(paramNames, part[1:])
			patternParts = append(patternParts, "([^/]+)")
		} else {
			patternParts = append(patternParts, regexp.QuoteMeta(part))