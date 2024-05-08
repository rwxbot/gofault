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