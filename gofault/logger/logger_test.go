package logger

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofault/gofault/core"
)

// mockLogger captures log output for testing.
type mockLogger struct {
	*bytes.Buffer
}

func (m *mockLogger) Debug(msg string, args ...any) {
	m.Write([]byte(fmt.Sprintf("DEBUG: "+msg+"\n", args...)))
}
func (m *mockLogger) Info(msg string, args ...any) {
	m.Write([]byte(fmt.Sprintf("INFO: "+msg+"\n", args...)))
}
func (m *mockLogger) Warn(msg string, args ...any) {
	m.Write([]byte(fmt.Sprintf("WARN: "+msg+"\n", args...)))
}