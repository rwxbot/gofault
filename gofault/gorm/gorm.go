// Package gorm provides GORM database integration for the gofault framework.
package gorm

import (
	"fmt"
	"time"

	"github.com/gofault/gofault/core"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Dialect supported database drivers.
type Dialect string

const (
	DialectMySQL    Dialect = "mysql"
	DialectPostgres Dialect = "postgres"
	DialectSQLite   Dialect = "sqlite"
)

// Config holds database configuration.
type Config struct {
	// Dialect is the database driver (mysql, postgres, sqlite).
	Dialect Dialect
	// DSN is the data source name (connection string).
	DSN string
	// MaxOpenConns sets the maximum number of open connections.
	MaxOpenConns int
	// MaxIdleConns sets the maximum number of idle connections.
	MaxIdleConns int
	// ConnMaxLifetime sets the maximum lifetime of a connection.
	ConnMaxLifetime int // in seconds
	// LogLevel sets GORM log level (default: logger.Warn).
	LogLevel logger.LogLevel
	// Silent suppresses all GORM logs when true.
	Silent bool
}

// DefaultConfig returns a default configuration.
func DefaultConfig() Config {
	return Config{
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: 3600,
		LogLevel:        logger.Warn,
		Silent:          false,
	}
}