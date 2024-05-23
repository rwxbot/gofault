// Package config provides configuration management for gofault.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds application configuration.
type Config struct {
	App      AppConfig      `yaml:"app"`