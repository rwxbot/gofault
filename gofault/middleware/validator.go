package middleware

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

// ValidationError represents a single validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var msgs []string
	for _, e := range ve {
		msgs = append(msgs, e.Field+": "+e.Message)
	}
	return strings.Join(msgs, "; ")
}

// ValidatorConfig holds configuration for the validator middleware.
type ValidatorConfig struct {
	// BindTarget is the target struct type to bind and validate.
	BindTarget interface{}
	// SkipMissing causes missing query/path params to be skipped instead of causing errors.
	SkipMissing bool
}

// ValidationRule defines a single validation rule.
type ValidationRule interface {
	Validate(field string, value interface{}) *ValidationError
}

// Required validates that a value is not empty.
type Required struct{}

func (r *Required) Validate(field string, value interface{}) *ValidationError {
	if value == nil || (reflect.TypeOf(value).Kind() == reflect.String && strings.TrimSpace(value.(string)) == "") {
		return &ValidationError{Field: field, Message: "is required"}
	}
	return nil
}

// MinLength validates minimum string length.
type MinLength struct {
	Length int
}

func (r *MinLength) Validate(field string, value interface{}) *ValidationError {
	if str, ok := value.(string); ok {
		if len(str) < r.Length {
			return &ValidationError{Field: field, Message: "must be at least " + itoa(r.Length) + " characters"}
		}
	}
	return nil
}

// MaxLength validates maximum string length.
type MaxLength struct {
	Length int
}

func (r *MaxLength) Validate(field string, value interface{}) *ValidationError {
	if str, ok := value.(string); ok {
		if len(str) > r.Length {
			return &ValidationError{Field: field, Message: "must be at most " + itoa(r.Length) + " characters"}
		}
	}
	return nil
}

// Regex validates that a string matches a regex pattern.
type Regex struct {
	Pattern string