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