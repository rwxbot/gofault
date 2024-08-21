package middleware

import (
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

func TestRequired(t *testing.T) {
	r := &Required{}

	// Valid case
	err := r.Validate("field", "value")
	if err != nil {
		t.Errorf("expected no error for non-empty value, got %v", err)
	}

	// Empty string
	err = r.Validate("field", "")
	if err == nil {
		t.Error("expected error for empty string")
	}

	// Nil
	err = r.Validate("field", nil)
	if err == nil {
		t.Error("expected error for nil")
	}
}

func TestMinLength(t *testing.T) {
	r := &MinLength{Length: 5}

	err := r.Validate("field", "hello")
	if err != nil {
		t.Errorf("expected no error for length 5, got %v", err)
	}

	err = r.Validate("field", "hi")
	if err == nil {
		t.Error("expected error for length 2")
	}
}

func TestMaxLength(t *testing.T) {
	r := &MaxLength{Length: 5}

	err := r.Validate("field", "hello")
	if err != nil {
		t.Errorf("expected no error for length 5, got %v", err)
	}

	err = r.Validate("field", "hello world")
	if err == nil {
		t.Error("expected error for length > 5")
	}
}

func TestRegex(t *testing.T) {
	r := &Regex{Pattern: `^\d{3}-\d{4}$`}

	err := r.Validate("field", "123-4567")
	if err != nil {
		t.Errorf("expected no error for matching pattern, got %v", err)
	}

	err = r.Validate("field", "abc-defg")
	if err == nil {
		t.Error("expected error for non-matching pattern")
	}
}

func TestEmail(t *testing.T) {
	r := &Email{}

	err := r.Validate("field", "test@example.com")
	if err != nil {
		t.Errorf("expected no error for valid email, got %v", err)
	}

	err = r.Validate("field", "invalid-email")
	if err == nil {
		t.Error("expected error for invalid email")
	}
}

// Path/query parameters always arrive as strings; Min must parse them rather
// than silently skipping (the old `case string: // Skip` made every rule a no-op).
func TestMin_StringParam(t *testing.T) {
	r := &Min{Value: 10}
	if err := r.Validate("age", "5"); err == nil {
		t.Error("string value below minimum should fail validation")
	}
	if err := r.Validate("age", "15"); err != nil {
		t.Errorf("string value above minimum should pass, got %v", err)
	}
	if err := r.Validate("age", "abc"); err == nil {
		t.Error("non-numeric string should fail validation, not pass silently")
	}
}

func TestMax_StringParam(t *testing.T) {
	r := &Max{Value: 100}
	if err := r.Validate("page", "150"); err == nil {
		t.Error("string value above maximum should fail validation")
	}
	if err := r.Validate("page", "50"); err != nil {
		t.Errorf("string value below maximum should pass, got %v", err)
	}
	if err := r.Validate("page", "xyz"); err == nil {
		t.Error("non-numeric string should fail validation, not pass silently")
	}
}

// A bad pattern must not panic the request path (MustCompile used to run
// inside Validate), and it must fail closed rather than pass everything.
func TestRegex_InvalidPatternFailsClosed(t *testing.T) {
	r := &Regex{Pattern: "([unclosed"}
	err := r.Validate("code", "anything")
	if err == nil {
		t.Fatal("invalid pattern should produce a validation error, not a pass")
	}
	// And it must not panic even when hammered concurrently.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 20; i++ {
			_ = r.Validate("code", "x")
		}
	}()
	for i := 0; i < 20; i++ {
		_ = r.Validate("code", "x")
	}
	<-done
}

// Concurrent first-use of a valid pattern must not race on the lazy compile.
func TestRegex_ConcurrentCompile(t *testing.T) {
	r := &Regex{Pattern: `^\d{3}$`}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := r.Validate("code", "123"); err != nil {
				t.Errorf("valid value rejected: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestMin(t *testing.T) {
	r := &Min{Value: 10}

	err := r.Validate("field", 15)
	if err != nil {
		t.Errorf("expected no error for value 15, got %v", err)
	}

	err = r.Validate("field", 5)
	if err == nil {
		t.Error("expected error for value 5")
	}
}

func TestMax(t *testing.T) {
	r := &Max{Value: 100}

	err := r.Validate("field", 50)
	if err != nil {
		t.Errorf("expected no error for value 50, got %v", err)
	}

	err = r.Validate("field", 150)
	if err == nil {
		t.Error("expected error for value 150")
	}
}

func TestValidationErrors_Error(t *testing.T) {
	errs := ValidationErrors{
		{Field: "name", Message: "is required"},
		{Field: "email", Message: "invalid format"},
	}

	str := errs.Error()
	if str != "name: is required; email: invalid format" {
		t.Errorf("unexpected error string: %s", str)
	}
}

func TestValidateRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?name=John&age=25", nil)
	ctx := &core.Ctx{Request: req, Params: make(map[string]string)}

	rules := Rules{