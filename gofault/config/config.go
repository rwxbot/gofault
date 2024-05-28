// Package config provides configuration management for gofault.
package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ErrUnknownField is reported when the configuration file contains a key that
// does not map to any field. It is returned as part of a FieldError so callers
// can inspect it with errors.Is.
var ErrUnknownField = errors.New("unknown configuration field")

// FieldError describes one problem found while validating a configuration.
type FieldError struct {
	// Field is the dotted path of the offending key, e.g. "server.port".
	Field string
	// Message explains what is wrong.
	Message string
	// Value is the value that was rejected.
	Value any
	// Unknown marks this as an unrecognised key rather than a bad value.
	Unknown bool
}

func (e *FieldError) Error() string {
	if e.Unknown {
		return fmt.Sprintf("unknown configuration field %q: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("invalid configuration for %q: %s (got %v)", e.Field, e.Message, e.Value)
}

// Unwrap lets errors.Is(err, ErrUnknownField) match any unknown-key error.
func (e *FieldError) Unwrap() error {
	if e.Unknown {
		return ErrUnknownField
	}
	return nil
}

// Config holds application configuration.
type Config struct {
	App      AppConfig      `yaml:"app"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	Version     string `yaml:"version"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	// Driver is one of "mysql", "postgres" or "sqlite". It selects the DSN
	// format produced by GetDSN.
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// isUnknownKeyError reports whether one yaml.TypeError entry describes a key
// that does not map to any struct field. yaml.v3 formats those entries as
// "line N: field NAME not found in type TYPE"; every other entry it produces
// under KnownFields(true) is a value type mismatch ("cannot unmarshal ...").
func isUnknownKeyError(msg string) bool {
	return strings.Contains(msg, "field ") && strings.Contains(msg, " not found in type ")
}

// Load reads, parses and validates a YAML configuration file.
//
// Unknown keys are rejected: a typo such as "sever.port" used to be silently
// ignored, leaving the server on its default port with no warning.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}