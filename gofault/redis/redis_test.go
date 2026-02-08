package redis

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()