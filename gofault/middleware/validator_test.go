package middleware

import (
	"net/http/httptest"
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
		"name": {&Required{}, &MinLength{Length: 2}},
		"age":  {&Required{}},
	}

	err := ValidateRequest(ctx, rules)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateRequest_MissingField(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := &core.Ctx{Request: req, Params: make(map[string]string)}

	rules := Rules{
		"name": {&Required{}},
	}

	err := ValidateRequest(ctx, rules)
	if err == nil {
		t.Error("expected error for missing required field")
	}
	if _, ok := err.(exception.HTTPException); !ok {
		t.Error("expected HTTPException")
	}
}

func TestValidateRequest_InvalidEmail(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?email=not-an-email", nil)
	ctx := &core.Ctx{Request: req, Params: make(map[string]string)}

	rules := Rules{
		"email": {&Required{}, &Email{}},
	}

	err := ValidateRequest(ctx, rules)
	if err == nil {
		t.Error("expected error for invalid email")
	}
}
