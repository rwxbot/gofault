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

// Database is the core database module.
type Database struct {
	core.Module
	DB     *gorm.DB
	Config Config
}

// NewDatabase creates a new Database module.
func NewDatabase(name string, config Config) (*Database, error) {
	db, err := openDB(config)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	d := &Database{
		Module: *core.NewModule(name),
		DB:     db,
		Config: config,
	}

	d.RegisterOnShutdown(&gormShutdown{db})

	return d, nil
}

// MustNewDatabase creates a new Database module and panics on error.
func MustNewDatabase(name string, config Config) *Database {
	d, err := NewDatabase(name, config)
	if err != nil {
		panic(err)
	}
	return d
}

// GetDB returns the underlying GORM DB instance.
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}
