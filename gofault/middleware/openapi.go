package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/gofault/gofault/core"
)

// OpenAPIConfig holds configuration for the OpenAPI documentation middleware.
type OpenAPIConfig struct {
	// Enabled enables the OpenAPI middleware.
	Enabled bool
	// Path is the URL path for the OpenAPI endpoint.
	Path string
	// Title is the API title.
	Title string
	// Version is the API version.
	Version string
	// Description is the API description.
	Description string
}

// DefaultOpenAPIConfig returns a default OpenAPI configuration.
func DefaultOpenAPIConfig() OpenAPIConfig {
	return OpenAPIConfig{
		Enabled:     true,
		Path:        "/openapi.json",
		Title:       "API",