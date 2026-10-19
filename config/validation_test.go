package config

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

// A typo such as "sever.port" used to be ignored silently, leaving the server
// on its default port.
func TestLoad_RejectsUnknownField(t *testing.T) {
	path := writeConfig(t, `
app:
  name: demo
sever:
  port: 8080
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an error for an unknown field")
	}
	if !errors.Is(err, ErrUnknownField) {
		t.Errorf("error = %v, want it to match ErrUnknownField", err)
	}
	if !strings.Contains(err.Error(), "sever") {
		t.Errorf("error should name the offending key, got %v", err)
	}
}

// A value type mismatch (a string where an int is expected) also surfaces as a
// *yaml.TypeError under KnownFields(true), but it is not an unknown key: it
// must not be labelled Unknown and must not match ErrUnknownField.
func TestLoad_TypeMismatchIsNotUnknownField(t *testing.T) {
	path := writeConfig(t, `
server:
  port: "notanint"
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an error for a type mismatch")
	}
	if errors.Is(err, ErrUnknownField) {
		t.Errorf("a type mismatch must not match ErrUnknownField, got %v", err)
	}
	var fe *FieldError
	if !errors.As(err, &fe) {
		t.Fatalf("error = %v (%T), want *FieldError", err, err)
	}
	if fe.Unknown {
		t.Errorf("type mismatch marked Unknown=true: %+v", fe)
	}
	if !strings.Contains(err.Error(), "notanint") {
		t.Errorf("error should quote the bad value, got %v", err)
	}
}

// When a file has both an unknown key and a type mismatch, the unknown-key
// signal must still win so errors.Is(err, ErrUnknownField) matches.
func TestLoad_MixedErrorsStillMatchUnknownField(t *testing.T) {
	path := writeConfig(t, `
server:
  port: "abc"
app:
  boguskey: 1
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !errors.Is(err, ErrUnknownField) {
		t.Errorf("mixed errors should match ErrUnknownField, got %v", err)
	}
}

func TestLoad_RejectsUnknownNestedField(t *testing.T) {
	path := writeConfig(t, `
app:
  name: demo
  unexpected: x
`)

	if _, err := Load(path); err == nil {
		t.Fatal("expected an error for an unknown nested field")
	}
}

func TestValidate_PortRange(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"valid high port", Config{Server: ServerConfig{Port: 65535}}, false},
		{"zero port is allowed", Config{Server: ServerConfig{Port: 0}}, false},
		{"negative port", Config{Server: ServerConfig{Port: -1}}, true},
		{"port above range", Config{Server: ServerConfig{Port: 70000}}, true},
		{"negative db port", Config{Database: DatabaseConfig{Port: -5}}, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.cfg.Validate()
			if c.wantErr && err == nil {
				t.Error("expected a validation error")
			}
			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidate_Driver(t *testing.T) {
	valid := []string{"", "mysql", "postgres", "postgresql", "sqlite", "sqlite3", "MySQL"}
	for _, d := range valid {
		cfg := Config{Database: DatabaseConfig{Driver: d}}
		if err := cfg.Validate(); err != nil {
			t.Errorf("driver %q rejected: %v", d, err)
		}
	}

	cfg := Config{Database: DatabaseConfig{Driver: "oracle"}}
	if err := cfg.Validate(); err == nil {
		t.Error("expected an unknown driver to be rejected")
	}
}

func TestValidate_LogSettings(t *testing.T) {
	bad := Config{Log: LogConfig{Level: "verbose", Format: "xml"}}
	err := bad.Validate()
	if err == nil {
		t.Fatal("expected invalid log settings to be rejected")
	}
	// Both problems should be reported, not just the first.
	if !strings.Contains(err.Error(), "log.level") || !strings.Contains(err.Error(), "log.format") {
		t.Errorf("expected both log fields reported, got %v", err)
	}

	ok := Config{Log: LogConfig{Level: "debug", Format: "json"}}
	if err := ok.Validate(); err != nil {
		t.Errorf("valid log settings rejected: %v", err)
	}
}

// GetDSN used to always emit the MySQL form, so a PostgreSQL or SQLite
// deployment got a DSN its driver could not parse.
func TestGetDSN_DriverAware(t *testing.T) {
	mysql := DatabaseConfig{
		Driver: "mysql", Host: "localhost", Port: 3306,
		Name: "mydb", User: "user", Password: "pass",
	}
	got, err := mysql.GetDSN()
	if err != nil {
		t.Fatalf("mysql: %v", err)
	}
	if got != "user:pass@tcp(localhost:3306)/mydb?charset=utf8mb4" {
		t.Errorf("mysql DSN = %q", got)
	}

	pg := DatabaseConfig{
		Driver: "postgres", Host: "localhost", Port: 5432,
		Name: "app", User: "user", Password: "pass",
	}
	got, err = pg.GetDSN()
	if err != nil {
		t.Fatalf("postgres: %v", err)
	}
	if !strings.HasPrefix(got, "postgres://") {
		t.Errorf("postgres DSN = %q, want a postgres:// URL", got)
	}
	if strings.Contains(got, "@tcp(") {
		t.Errorf("postgres DSN still uses the mysql form: %q", got)
	}

	sqlite := DatabaseConfig{Driver: "sqlite", Name: "/var/db/app.db"}
	got, err = sqlite.GetDSN()
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if got != "/var/db/app.db" {
		t.Errorf("sqlite DSN = %q, want the file path", got)
	}

	// An unknown driver must report an error rather than a broken DSN.
	bad := DatabaseConfig{Driver: "oracle", Name: "x"}
	if _, err := bad.GetDSN(); err == nil {
		t.Error("expected an error for an unknown driver")
	}

	// SQLite needs a path.
	empty := DatabaseConfig{Driver: "sqlite"}
	if _, err := empty.GetDSN(); err == nil {
		t.Error("expected an error for sqlite with no name")
	}
}

// A generated secret full of '@', ':', '#' or spaces used to corrupt the
// credential split; the mysql driver takes the LAST '@' as the separator and
// the FIRST ':' inside the userinfo, so those characters must pass through
// unescaped. The database name is the part that does need escaping: the
// driver treats '?' and '/' as structure and decodes %XX escapes there.
func TestGetDSN_MySQLSpecialCharacters(t *testing.T) {
	db := DatabaseConfig{
		Driver: "mysql", Host: "127.0.0.1", Port: 3306,
		Name: "db", User: "app", Password: "p@ss:w#d &x",
	}
	got, err := db.GetDSN()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "app:p@ss:w#d &x@tcp(127.0.0.1:3306)/db?charset=utf8mb4"
	if got != want {
		t.Errorf("DSN = %q, want %q", got, want)
	}

	// A name with structural characters must be percent-encoded so the
	// driver's own unescape turns it back into the original name.
	db.Name = "d/b?x%y"
	got, err = db.GetDSN()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want = "app:p@ss:w#d &x@tcp(127.0.0.1:3306)/d%2Fb%3Fx%25y?charset=utf8mb4"
	if got != want {
		t.Errorf("DSN = %q, want %q", got, want)
	}

	// ':' in the username cannot be represented: the driver splits the
	// userinfo at the first colon and does not decode escapes there.
	bad := DatabaseConfig{
		Driver: "mysql", Host: "h", Port: 3306,
		Name: "db", User: "do:main", Password: "p",
	}
	if _, err := bad.GetDSN(); err == nil {
		t.Error("expected an error for a mysql user containing ':'")
	}
	if got := bad.GetMySQLDSN(); got != "" {
		t.Errorf("GetMySQLDSN = %q, want empty for an unrepresentable user", got)
	}
}

// The postgres DSN is a URL, so the userinfo and the database name must be
// percent-encoded and must decode back to the original values.
func TestGetDSN_PostgresEscaping(t *testing.T) {
	db := DatabaseConfig{
		Driver: "postgres", Host: "db.internal", Port: 5432,
		Name: "app/db?x", User: "ad:min", Password: "p@ss w#1",
	}
	got, err := db.GetDSN()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("postgres DSN %q does not parse as a URL: %v", got, err)
	}
	if u.User == nil {
		t.Fatalf("postgres DSN %q has no userinfo", got)
	}
	pass, ok := u.User.Password()
	if !ok {
		t.Fatalf("postgres DSN %q has no password", got)
	}
	if u.User.Username() != db.User {
		t.Errorf("user decoded to %q, want %q", u.User.Username(), db.User)
	}
	if pass != db.Password {
		t.Errorf("password decoded to %q, want %q", pass, db.Password)
	}
	if name := strings.TrimPrefix(u.Path, "/"); name != db.Name {
		t.Errorf("database decoded to %q, want %q", name, db.Name)
	}
	// The raw DSN must not carry a second '@' that would move the boundary.
	if strings.Count(strings.TrimPrefix(got, "postgres://"), "@") != 1 {
		t.Errorf("postgres DSN %q has an unescaped '@' in the credentials", got)
	}
}

// An empty driver keeps the historical MySQL behaviour.
func TestGetDSN_DefaultsToMySQL(t *testing.T) {
	db := DatabaseConfig{Host: "h", Port: 3306, Name: "n", User: "u", Password: "p"}
	got, err := db.GetDSN()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "@tcp(") {
		t.Errorf("DSN = %q, want the mysql form when no driver is set", got)
	}
	if db.GetMySQLDSN() != got {
		t.Error("GetMySQLDSN should match the default GetDSN")
	}
}

func TestGetAddress(t *testing.T) {
	// net.JoinHostPort brackets IPv6 literals.
	c := ServerConfig{Host: "::1", Port: 8080}
	if got := c.GetAddress(); got != "[::1]:8080" {
		t.Errorf("GetAddress() = %q, want [::1]:8080", got)
	}

	c2 := ServerConfig{Host: "localhost", Port: 8080}
	if got := c2.GetAddress(); got != "localhost:8080" {
		t.Errorf("GetAddress() = %q, want localhost:8080", got)
	}
}

// Load must run validation, so a bad port is caught at startup.
func TestLoad_Validates(t *testing.T) {
	path := writeConfig(t, `
server:
  port: -1
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected Load to reject a negative port")
	}

	good := writeConfig(t, `
app:
  name: demo
server:
  host: localhost
  port: 8080
log:
  level: info
  format: json
`)
	cfg, err := Load(good)
	if err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	if cfg.App.Name != "demo" || cfg.Server.Port != 8080 {
		t.Errorf("config not parsed correctly: %+v", cfg)
	}
}

func TestFieldError_Unwrap(t *testing.T) {
	unknown := &FieldError{Field: "a", Message: "b", Unknown: true}
	if !errors.Is(unknown, ErrUnknownField) {
		t.Error("an unknown-field error should match ErrUnknownField")
	}

	plain := &FieldError{Field: "server.port", Message: "bad", Value: -1}
	if errors.Is(plain, ErrUnknownField) {
		t.Error("a value error should not match ErrUnknownField")
	}
	if !strings.Contains(plain.Error(), "server.port") {
		t.Errorf("error should name the field, got %v", plain)
	}
}
