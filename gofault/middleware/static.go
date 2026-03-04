package middleware

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofault/gofault/core"
)

// StaticConfig holds configuration for the static file middleware.