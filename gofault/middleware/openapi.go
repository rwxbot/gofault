package middleware

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gofault/gofault/core"
)

// OpenAPIConfig holds configuration for the OpenAPI documentation middleware.