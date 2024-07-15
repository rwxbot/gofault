package controllers

import (
	"github.com/gofault/gofault/core"
)

// HelloController handles hello-related routes.
type HelloController struct{}

func NewHelloController() *HelloController {
	return &HelloController{}
}

// Routes returns the routes for this controller.
func (c *HelloController) Routes() []core.Route {
	return []core.Route{
		{Method: "GET", Path: "/", Handler: "Index"},
		{Method: "GET", Path: "/hello/:name", Handler: "Greet"},
	}
}

// Prefix returns the route prefix for this controller.
func (c *HelloController) Prefix() string {
	return "/api/v1"
}

// Index handles GET /
func (c *HelloController) Index(ctx *core.Ctx) error {