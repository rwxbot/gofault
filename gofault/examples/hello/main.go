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