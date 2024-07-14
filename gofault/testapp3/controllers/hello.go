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