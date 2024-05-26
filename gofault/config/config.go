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
