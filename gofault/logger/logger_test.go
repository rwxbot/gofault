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
func (m *mockLogger) Error(msg string, args ...any) {
	m.Write([]byte(fmt.Sprintf("ERROR: "+msg+"\n", args...)))
}

func TestNew(t *testing.T) {
	l := New("[TEST] ", LevelInfo)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
	// Should not panic
	l.Debug("debug")
	l.Info("info")
	l.Warn("warn")
	l.Error("error")
}

func TestDefaultLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := &defaultLogger{
		logger: log.New(&buf, "", 0),
		level:  LevelWarn,
	}

	l.Debug("should not appear")
	l.Info("should not appear")
	l.Warn("should appear")
	l.Error("should appear")

	output := buf.String()
	if strings.Contains(output, "DEBUG") {
		t.Error("DEBUG should be filtered")
	}
	if strings.Contains(output, "INFO") {
		t.Error("INFO should be filtered")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("WARN should appear")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("ERROR should appear")
	}
}

func TestDefaultLogger_FormattedMessage(t *testing.T) {
	var buf bytes.Buffer
	l := &defaultLogger{
		logger: log.New(&buf, "", 0),
		level:  LevelDebug,
	}

	l.Info("user=%s id=%d", "alice", 42)

	output := buf.String()
	if !strings.Contains(output, "user=alice") {
		t.Errorf("expected formatted output, got: %s", output)
	}
	if !strings.Contains(output, "id=42") {
		t.Errorf("expected formatted output, got: %s", output)
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    int
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{99, "UNKNOWN"},
	}

	for _, tt := range tests {
		got := levelString(tt.level)
		if got != tt.expected {
			t.Errorf("levelString(%d) = %s, want %s", tt.level, got, tt.expected)
		}
	}
}

func TestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	ml := &mockLogger{Buffer: &buf}

	middleware := LoggingMiddleware(ml)

	var nextCalled bool
	next := func(ctx *core.Ctx) error {
		nextCalled = true
		ctx.StatusCode = http.StatusCreated
		return nil
	}

	req := httptest.NewRequest("POST", "/api/users/123", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)
	ctx.Request = req

	err := middleware(ctx, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !nextCalled {
		t.Error("next handler was not called")
	}
	if ctx.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", ctx.StatusCode, http.StatusCreated)
	}

	output := buf.String()
	if !strings.Contains(output, "POST") {
		t.Errorf("expected POST method in log, got: %s", output)
	}
	if !strings.Contains(output, "/api/users/123") {
		t.Errorf("expected path in log, got: %s", output)
	}
	if !strings.Contains(output, "201") {
		t.Errorf("expected status code 201 in log, got: %s", output)
	}
}

func TestLoggingMiddleware_ErrorPropagation(t *testing.T) {
	var buf bytes.Buffer
	ml := &mockLogger{Buffer: &buf}

	middleware := LoggingMiddleware(ml)

	expectedErr := fmt.Errorf("handler error")
	next := func(ctx *core.Ctx) error {
		return expectedErr
	}

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	err := middleware(ctx, next)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestLoggingMiddleware_LogsDuration(t *testing.T) {
	var buf bytes.Buffer
	ml := &mockLogger{Buffer: &buf}

	middleware := LoggingMiddleware(ml)

	next := func(ctx *core.Ctx) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	output := buf.String()
	// Duration should be logged (format: "POST /path status duration")
	if !strings.Contains(output, "0s") && !strings.Contains(output, "ns") {
		t.Errorf("expected duration in log output, got: %s", output)
	}
}

func TestLoggerInterface(t *testing.T) {
	// Verify Logger interface is implemented correctly
	var _ Logger = New("[TEST] ", LevelInfo)
}

func TestLoggingMiddleware_WithCustomLogger(t *testing.T) {
	var buf bytes.Buffer
	customLogger := &defaultLogger{
		logger: log.New(&buf, "", 0),
		level:  LevelInfo,
	}

	middleware := LoggingMiddleware(customLogger)

	next := func(ctx *core.Ctx) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected INFO level in log, got: %s", output)
	}
	if !strings.Contains(output, "GET") {
		t.Errorf("expected GET method in log, got: %s", output)
	}
}
