package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"