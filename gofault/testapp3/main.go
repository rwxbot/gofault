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

	// Register module
	if err := app.RegisterModules(mod); err != nil {
		log.Fatalf("Failed to register module: %v", err)
	}

	// Start server
	fmt.Println("Server starting on :8080...")
	if err := app.Start(8080); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
