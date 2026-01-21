// Package redis provides Redis integration for the gofault framework.
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/middleware"
	"github.com/redis/go-redis/v9"
)

// Config holds Redis configuration.