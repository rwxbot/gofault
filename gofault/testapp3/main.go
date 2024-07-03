package main

import (
	"fmt"
	"log"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/module"
	"github.com/gofault/gofault/testapp3/controllers"
)

func main() {
	// Create application
	app := module.New()

	// Create module
	mod := core.NewModule("testapp3")
	mod.RegisterControllers(controllers.NewHelloController())