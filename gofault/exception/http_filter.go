package exception

import (
	"encoding/json"

	"github.com/gofault/gofault/core"
)

// HTTPExceptionResponse is the JSON structure returned on HTTP exceptions.
type HTTPExceptionResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HTTPExceptionFilter is a filter that handles HTTPException errors.