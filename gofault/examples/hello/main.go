// Package main is a minimal example demonstrating gofault application structure.
package main

import (
	"log"

	"github.com/gofault/gofault/controller"
	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
	"github.com/gofault/gofault/module"
	"github.com/gofault/gofault/router"
)

// GreetingService is a simple injectable provider.
type GreetingService struct {
	prefix string
}

func NewGreetingService(prefix string) *GreetingService {
	return &GreetingService{prefix: prefix}
}

func (s *GreetingService) Provide() any { return s }

func (s *GreetingService) Greet(name string) string {
	return s.prefix + ", " + name + "!"
}

// LifecycleHook demonstrates OnBoot/OnShutdown lifecycle.
type LifecycleHook struct {
	name string
}

func NewLifecycleHook(name string) *LifecycleHook {
	return &LifecycleHook{name: name}
}

func (h *LifecycleHook) OnBoot() error {
	log.Printf("[Lifecycle] OnBoot: %s started", h.name)
	return nil
}

func (h *LifecycleHook) OnShutdown() error {
	log.Printf("[Lifecycle] OnShutdown: %s stopped", h.name)
	return nil
}

// GreetingController is the HTTP entry point for greeting endpoints.
type GreetingController struct {
	controller.BaseController
	svc *GreetingService
}

func NewGreetingController(svc *GreetingService) *GreetingController {
	return &GreetingController{svc: svc}
}

func (c *GreetingController) Prefix() string { return "/hello" }

func (c *GreetingController) Routes() []core.Route {
	return []core.Route{
		{Method: "GET", Path: "/greet/:name", Handler: "Greet"},
		{Method: "GET", Path: "/", Handler: "Index"},
		{Method: "GET", Path: "/error", Handler: "Error"},
	}
}

func (c *GreetingController) Greet(ctx *core.Ctx) error {
	name := controller.Param(ctx, "name")
	msg := c.svc.Greet(name)
	return controller.OK(ctx.Response, map[string]string{"message": msg})
}

func (c *GreetingController) Index(ctx *core.Ctx) error {
	return controller.OK(ctx.Response, map[string]string{"message": "gofault is running"})
}