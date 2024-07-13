package controllers

import (
	"github.com/gofault/gofault/core"
)

// HelloController handles hello-related routes.
type HelloController struct{}

func NewHelloController() *HelloController {
	return &HelloController{}