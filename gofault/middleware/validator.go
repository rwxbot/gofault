package middleware

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

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
//
// The pattern is compiled on first use under a mutex: MustCompile on the
// request path would panic the whole process on a bad pattern (rules are
// public API and can come from config), and the lazy assignment was a data
// race because one rule instance is shared by concurrent requests. A pattern
// that fails to compile fails closed: every value is rejected with a
// validation error rather than silently passing or panicking.
type Regex struct {
	Pattern string

	mu         sync.Mutex
	compiled   *regexp.Regexp
	compileErr error
}

func (r *Regex) compile() (*regexp.Regexp, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.compiled == nil && r.compileErr == nil {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			r.compileErr = err
		} else {
			r.compiled = re
		}
	}
	return r.compiled, r.compileErr
}

func (r *Regex) Validate(field string, value interface{}) *ValidationError {
	if str, ok := value.(string); ok {
		re, err := r.compile()
		if err != nil {
			return &ValidationError{Field: field, Message: "cannot validate against invalid pattern"}
		}
		if !re.MatchString(str) {
			return &ValidationError{Field: field, Message: "has invalid format"}
		}
	}
	return nil
}

// Email validates email format.
type Email struct{}

func (r *Email) Validate(field string, value interface{}) *ValidationError {
	if str, ok := value.(string); ok {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(str) {
			return &ValidationError{Field: field, Message: "must be a valid email address"}
		}
	}
	return nil
}

// Min validates minimum numeric value.
type Min struct {
	Value float64
}

func (r *Min) Validate(field string, value interface{}) *ValidationError {
	switch v := value.(type) {
	case int:
		if float64(v) < r.Value {
			return &ValidationError{Field: field, Message: "must be at least " + itoa(int(r.Value))}
		}
	case float64:
		if v < r.Value {
			return &ValidationError{Field: field, Message: "must be at least " + ftos(r.Value)}
		}
	case string:
		// Path and query parameters always arrive as strings, so skipping
		// them made every Min rule on a request field a silent no-op. Parse
		// and enforce; a non-numeric value is a validation failure, not a
		// pass.
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return &ValidationError{Field: field, Message: "must be a number of at least " + ftos(r.Value)}
		}
		if n < r.Value {
			return &ValidationError{Field: field, Message: "must be at least " + ftos(r.Value)}
		}
	}
	return nil
}

// Max validates maximum numeric value.
type Max struct {
	Value float64
}

func (r *Max) Validate(field string, value interface{}) *ValidationError {
	switch v := value.(type) {
	case int:
		if float64(v) > r.Value {
			return &ValidationError{Field: field, Message: "must be at most " + itoa(int(r.Value))}
		}
	case float64:
		if v > r.Value {
			return &ValidationError{Field: field, Message: "must be at most " + ftos(r.Value)}
		}
	case string:
		// Same as Min: request parameters are strings, and skipping them
		// silently disabled every Max rule on a request field.
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return &ValidationError{Field: field, Message: "must be a number of at most " + ftos(r.Value)}
		}
		if n > r.Value {
			return &ValidationError{Field: field, Message: "must be at most " + ftos(r.Value)}
		}
	}
	return nil
}

// Rules defines validation rules for a struct field.
type Rules map[string][]ValidationRule

// ValidateRequest validates the request against defined rules.
//
// Rules are supplied per call rather than per route; there is no validator
// middleware. A previous ValidatorConfig (with BindTarget and SkipMissing)
// promised one but was never read by any code, so setting it did nothing.
func ValidateRequest(ctx *core.Ctx, rules Rules) error {
	var errs ValidationErrors

	for field, rules := range rules {
		value := getFieldValue(ctx, field)
		for _, rule := range rules {
			if err := rule.Validate(field, value); err != nil {
				errs = append(errs, *err)
			}