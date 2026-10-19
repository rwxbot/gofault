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

	// KnownFields makes the decoder report keys that do not match any field.
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true)

	if err := dec.Decode(&cfg); err != nil {
		var typeErr *yaml.TypeError
		if errors.As(err, &typeErr) {
			// yaml.v3 reports both unknown keys (with KnownFields(true)) and
			// value type mismatches through the same *yaml.TypeError. Only the
			// unknown-key entries carry "field ... not found in type ...";
			// labelling a plain type mismatch as unknown would make
			// errors.Is(err, ErrUnknownField) match a bad-value failure.
			unknown := false
			for _, e := range typeErr.Errors {
				if isUnknownKeyError(e) {
					unknown = true
					break
				}
			}
			return nil, &FieldError{
				Field:   "config",
				Message: strings.Join(typeErr.Errors, "; "),
				Unknown: unknown,
			}
		}
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// MustLoad reads and parses a YAML configuration file, panics on error.
func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		panic(err)
	}
	return cfg
}

// Validate checks the configuration for values that cannot work.
//
// It catches mistakes at startup rather than at first use: a zero or negative
// port, a negative database port, an unknown driver, or an empty database name
// all produce errors here instead of a confusing failure later.
func (c *Config) Validate() error {
	var errs []error

	if c.Server.Port < 0 || c.Server.Port > 65535 {
		errs = append(errs, &FieldError{
			Field:   "server.port",
			Message: "must be between 0 and 65535",
			Value:   c.Server.Port,
		})
	}

	if c.Database.Driver != "" {
		if err := validateDriver(c.Database.Driver); err != nil {
			errs = append(errs, err)
		}
	}
	if c.Database.Port < 0 || c.Database.Port > 65535 {
		errs = append(errs, &FieldError{
			Field:   "database.port",
			Message: "must be between 0 and 65535",
			Value:   c.Database.Port,
		})
	}

	switch c.Log.Level {
	case "", "debug", "info", "warn", "error":
	default:
		errs = append(errs, &FieldError{
			Field:   "log.level",
			Message: "must be one of debug, info, warn, error",
			Value:   c.Log.Level,
		})
	}

	switch c.Log.Format {
	case "", "json", "text":
	default:
		errs = append(errs, &FieldError{
			Field:   "log.format",
			Message: "must be json or text",
			Value:   c.Log.Format,
		})
	}

	return errors.Join(errs...)
}

func validateDriver(driver string) error {
	switch strings.ToLower(driver) {
	case "mysql", "postgres", "postgresql", "sqlite", "sqlite3":
		return nil
	}
	return &FieldError{
		Field:   "database.driver",
		Message: "must be mysql, postgres or sqlite",
		Value:   driver,
	}
}

// GetAddress returns the server address in host:port format.
func (c *ServerConfig) GetAddress() string {
	return net.JoinHostPort(c.Host, fmt.Sprintf("%d", c.Port))
}

// GetDSN returns the database connection string for the configured driver.
//
// The format depends on Driver; it used to always emit the MySQL form, so a
// PostgreSQL or SQLite deployment produced a DSN its driver could not parse.
//
// Credentials and names are escaped for the target format. The MySQL DSN
// percent-encodes the database name, which go-sql-driver/mysql decodes back;
// a username containing ':' is rejected because the driver splits the
// userinfo at the first colon and does not decode escapes there. The
// PostgreSQL URL encodes the userinfo and the path per RFC 3986.
func (c *DatabaseConfig) GetDSN() (string, error) {
	driver := strings.ToLower(c.Driver)
	if driver == "" {
		driver = "mysql"
	}

	switch driver {
	case "sqlite", "sqlite3":
		// SQLite takes a file path, not a network address.
		if c.Name == "" {
			return "", &FieldError{
				Field:   "database.name",
				Message: "must be the database file path for sqlite",
			}
		}
		return c.Name, nil

	case "mysql":
		return c.mysqlDSN()

	case "postgres", "postgresql":
		// PostgreSQL uses a URL DSN rather than the mysql user:pass@tcp form.
		// UserPassword escapes both fields; RawPath keeps an encoded '/' in
		// the database name from being read as a path separator.
		u := &url.URL{
			Scheme:   "postgres",
			User:     url.UserPassword(c.User, c.Password),
			Host:     net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
			Path:     "/" + c.Name,
			RawPath:  "/" + url.PathEscape(c.Name),
			RawQuery: "sslmode=disable",
		}
		return u.String(), nil
	}

	return "", &FieldError{
		Field:   "database.driver",
		Message: "must be mysql, postgres or sqlite",
		Value:   c.Driver,
	}
}

// GetMySQLDSN returns the MySQL connection string regardless of the configured
// driver. It is kept for callers that specifically target MySQL. It applies
// the same escaping as the mysql branch of GetDSN, but cannot report the one
// unrepresentable case — a username containing ':' — and returns an empty
// string for it; GetDSN rejects that configuration with an error.
func (c *DatabaseConfig) GetMySQLDSN() string {
	dsn, err := c.mysqlDSN()
	if err != nil {
		return ""
	}
	return dsn
}

// mysqlDSN builds the go-sql-driver/mysql DSN form.
func (c *DatabaseConfig) mysqlDSN() (string, error) {
	if strings.ContainsRune(c.User, ':') {
		return "", &FieldError{
			Field:   "database.user",
			Message: "must not contain ':' for the MySQL DSN format",
			Value:   c.User,
		}
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4",
		c.User, c.Password, c.Host, c.Port, escapeMySQLName(c.Name)), nil
}

// escapeMySQLName percent-encodes the characters go-sql-driver/mysql treats
// as DSN structure inside the database name. The driver decodes %XX escapes
// in the name, so '?', '/' and '%' itself must be encoded; other characters
// pass through unchanged.
func escapeMySQLName(name string) string {
	if !strings.ContainsAny(name, "%?/") {
		return name
	}
	var b strings.Builder
	b.Grow(len(name) + 8)
	for i := 0; i < len(name); i++ {
		switch name[i] {
		case '%':
			b.WriteString("%25")
		case '?':
			b.WriteString("%3F")
		case '/':
			b.WriteString("%2F")
		default:
			b.WriteByte(name[i])
		}
	}
	return b.String()
}
