// Package gorm provides GORM database integration for the gofault framework.
package gorm

import (
	"fmt"
	"sync"
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