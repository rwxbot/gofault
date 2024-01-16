// Package exception provides HTTP exception types and exception filtering.
package exception

import (
	"fmt"
	"net/http"
)

// HTTPException is the base error type for HTTP-related exceptions.