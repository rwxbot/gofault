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
