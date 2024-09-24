package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gofault/gofault/core"
)

// RecoveryConfig holds configuration for the recovery middleware.