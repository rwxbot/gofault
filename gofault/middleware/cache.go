package middleware

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gofault/gofault/core"
)

// CacheConfig holds configuration for the cache middleware.