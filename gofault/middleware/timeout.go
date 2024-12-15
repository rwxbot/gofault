package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gofault/gofault/core"
)

// TimeoutConfig holds configuration for the timeout middleware.