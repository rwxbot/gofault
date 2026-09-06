// Package cmd provides CLI commands for gofault.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "new":
		if err := runNew(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println("gofault v1.0.0")
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`gofault - A lightweight Go web framework

Usage:
    gofault <command> [arguments]

Commands:
    new [name]    Create a new gofault project
    version       Print the version number
    help          Show this help message

Examples:
    gofault new myapp
    gofault version`)
}

func runNew() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("project name required\nUsage: gofault new <project-name>")
	}
	name := os.Args[2]
	return createProject(name)
}

func createProject(name string) error {
	// Determine working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	projectDir := filepath.Join(wd, name)

	// Create project directory
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Generate content from templates
	mainContent, err := executeTemplate(mainGoTemplate, map[string]string{"Name": name})