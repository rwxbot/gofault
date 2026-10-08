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
		Version:     "1.0.0",
		Description: "API Documentation",
	}
}

// OpenAPIMiddleware creates a middleware that serves OpenAPI specification.
func OpenAPIMiddleware(config OpenAPIConfig) core.Handler {
	if !config.Enabled {
		return nil
	}

	if config.Path == "" {
		config.Path = "/openapi.json"
	}

	spec := map[string]any{
		"openapi": "3.0.0",
		"info": map[string]any{
			"title":       config.Title,
			"version":     config.Version,
			"description": config.Description,
		},
		"paths": map[string]any{},
	}

	return func(ctx *core.Ctx) error {
		ctx.Response.Header().Set("Content-Type", "application/json")
		ctx.Response.WriteHeader(http.StatusOK)
		return json.NewEncoder(ctx.Response).Encode(spec)
	}
}

// OpenAPIDocument represents an OpenAPI 3.0 document.
type OpenAPIDocument struct {
	OpenAPI    string         `json:"openapi"`
	Info       map[string]any `json:"info"`
	Paths      map[string]any `json:"paths"`
	Components map[string]any `json:"components,omitempty"`
}

// AddPath adds a path to the OpenAPI document.
func (d *OpenAPIDocument) AddPath(path string, method string, operation map[string]any) {
	if d.Paths == nil {
		d.Paths = make(map[string]any)
	}
	if d.Paths[path] == nil {
		d.Paths[path] = make(map[string]any)
	}
	d.Paths[path].(map[string]any)[method] = operation
}

// NewOpenAPIDocument creates a new OpenAPI document.
func NewOpenAPIDocument(title, version, description string) *OpenAPIDocument {
	return &OpenAPIDocument{
		OpenAPI: "3.0.0",
		Info: map[string]any{
			"title":       title,
			"version":     version,
			"description": description,
		},
		Paths:      make(map[string]any),
		Components: make(map[string]any),
	}
}
