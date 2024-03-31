// Package controller provides the controller base type and HTTP method helpers.
package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gofault/gofault/core"
)

// BaseController provides helper methods for writing HTTP handlers.
type BaseController struct{}
