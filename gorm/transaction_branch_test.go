package gorm

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// The failure branches of Transaction (Begin error, Rollback error, Commit
// error) cannot be provoked through a real driver: sqlite/mysql/postgres
// always accept BEGIN and never fail COMMIT for a healthy connection. A fake
// dialector drives those branches deterministically.

// fakeDialector installs a scripted ConnPool without touching any real DB.
type fakeDialector struct {
	pool gorm.ConnPool
}

func (d fakeDialector) Name() string { return "fake" }

func (d fakeDialector) Initialize(db *gorm.DB) error {
	db.ConnPool = d.pool
	return nil
}

func (d fakeDialector) Migrator(db *gorm.DB) gorm.Migrator { return nil }

func (d fakeDialector) DataTypeOf(*schema.Field) string { return "" }

func (d fakeDialector) DefaultValueOf(*schema.Field) clause.Expression { return nil }

func (d fakeDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
}

func (d fakeDialector) QuoteTo(writer clause.Writer, str string) {}

func (d fakeDialector) Explain(sql string, vars ...interface{}) string { return sql }

// fakePool is the connection pool the fake dialector installs. Its BeginTx
// either fails (beginErr) or hands out the scripted fakeTx.
type fakePool struct {
	beginErr error
	tx       *fakeTx
}

func (p *fakePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errors.New("fake: PrepareContext not supported")
}

func (p *fakePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("fake: ExecContext not supported")
}

func (p *fakePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("fake: QueryContext not supported")
}

func (p *fakePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}

func (p *fakePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	if p.beginErr != nil {
		return nil, p.beginErr
	}
	return p.tx, nil
}

// fakeTx is the transaction handle: a ConnPool that also implements
// TxCommitter, with Commit/Rollback outcomes scripted per test.
type fakeTx struct {
	commitErr   error
	rollbackErr error
}

func (t *fakeTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errors.New("fake: PrepareContext not supported")
}

func (t *fakeTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("fake: ExecContext not supported")
}

func (t *fakeTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("fake: QueryContext not supported")
}

func (t *fakeTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}

func (t *fakeTx) Commit() error { return t.commitErr }

func (t *fakeTx) Rollback() error { return t.rollbackErr }

// openFake builds a *gorm.DB wired to the fake dialector. The fake pool does
// not implement Ping, so gorm.Open skips its automatic ping.
func openFake(t *testing.T, pool gorm.ConnPool) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(fakeDialector{pool: pool}, &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("open fake dialector: %v", err)
	}
	return db
}

// A Begin failure must surface as a wrapped "begin transaction" error, and
// fn must never run.
func TestTransaction_BeginError(t *testing.T) {
	beginErr := errors.New("fake: begin failed")
	db := openFake(t, &fakePool{beginErr: beginErr})

	called := false
	err := Transaction(db, func(tx *gorm.DB) error {
		called = true
		return nil
	})
	if err == nil {
		t.Fatal("expected an error when Begin fails")
	}
	if !strings.Contains(err.Error(), "begin transaction") {
		t.Errorf("error = %v, want it to mention begin transaction", err)
	}
	if !errors.Is(err, beginErr) {
		t.Errorf("error = %v, want it to wrap the begin error", err)
	}
	if called {
		t.Error("fn must not run when Begin fails")
	}
}

// When fn fails AND the rollback also fails, both errors must appear in the
// returned error: swallowing the rollback failure would hide a stuck open
// transaction.
func TestTransaction_RollbackErrorOnFnError(t *testing.T) {
	fnErr := errors.New("boom")
	rbErr := errors.New("fake: rollback failed")
	db := openFake(t, &fakePool{tx: &fakeTx{rollbackErr: rbErr}})

	err := Transaction(db, func(tx *gorm.DB) error {
		return fnErr
	})
	if err == nil {
		t.Fatal("expected an error when fn fails and rollback fails")
	}
	if !strings.Contains(err.Error(), "rollback:") {
		t.Errorf("error = %v, want it to mention the rollback failure", err)
	}
	if !strings.Contains(err.Error(), "original error") {
		t.Errorf("error = %v, want it to keep the original error visible", err)
	}
	if !errors.Is(err, fnErr) {
		t.Errorf("error = %v, want it to wrap the original fn error", err)
	}
}

// A failed Commit with a successful compensating Rollback must return the
// commit error wrapped, and the rollback must have been attempted.
func TestTransaction_CommitErrorRollbackOK(t *testing.T) {
	commitErr := errors.New("fake: commit failed")
	tx := &fakeTx{commitErr: commitErr, rollbackErr: nil}
	db := openFake(t, &fakePool{tx: tx})

	err := Transaction(db, func(tx *gorm.DB) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected an error when Commit fails")
	}
	if !strings.Contains(err.Error(), "commit:") {
		t.Errorf("error = %v, want it to mention the commit failure", err)
	}
	if !errors.Is(err, commitErr) {
		t.Errorf("error = %v, want it to wrap the commit error", err)
	}
}

// A failed Commit whose compensating Rollback also fails must report both
// failures rather than losing either.
func TestTransaction_CommitErrorRollbackAlsoFails(t *testing.T) {
	commitErr := errors.New("fake: commit failed")
	rbErr := errors.New("fake: rollback after failed commit failed")
	db := openFake(t, &fakePool{tx: &fakeTx{commitErr: commitErr, rollbackErr: rbErr}})

	err := Transaction(db, func(tx *gorm.DB) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected an error when Commit and Rollback both fail")
	}
	if !strings.Contains(err.Error(), "commit:") || !strings.Contains(err.Error(), "rollback:") {
		t.Errorf("error = %v, want it to mention both the commit and rollback failures", err)
	}
	if !errors.Is(err, rbErr) {
		t.Errorf("error = %v, want it to wrap the rollback error", err)
	}
}

// OnShutdown must return the error when the underlying *sql.DB handle cannot
// be obtained (the fake pool is neither *sql.DB nor a GetDBConnector).
func TestOnShutdown_DBHandleError(t *testing.T) {
	db := openFake(t, &fakePool{tx: &fakeTx{}})

	hook := &gormShutdown{db}
	if err := hook.OnShutdown(); err == nil {
		t.Fatal("expected OnShutdown to fail when the sql.DB handle is unavailable")
	}
}

// MustNewDatabase's success path returns the module; the existing panic tests
// only exercise the failure branch.
func TestMustNewDatabase_Success(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dialect = DialectSQLite
	cfg.DSN = ":memory:"
	cfg.Silent = true

	d := MustNewDatabase("must-ok", cfg)
	if d == nil || d.DB == nil {
		t.Fatal("MustNewDatabase returned nil database")
	}
	if d.Module.Name != "must-ok" {
		t.Errorf("name = %q, want %q", d.Module.Name, "must-ok")
	}
	sqlDB, err := d.DB.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Errorf("ping: %v", err)
	}
	_ = sqlDB.Close()
}

// The postgres case in openDB must be reachable: an unreachable postgres DSN
// must fail with a wrapped open error rather than an "unsupported dialect"
// error. Port 1 is reserved and never listening.
func TestOpenDB_PostgresUnreachable(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dialect = DialectPostgres
	cfg.DSN = "host=127.0.0.1 port=1 user=gofault sslmode=disable connect_timeout=1"
	cfg.Silent = true

	_, err := NewDatabase("pg-unreachable", cfg)
	if err == nil {
		t.Fatal("expected an error for an unreachable postgres")
	}
	if strings.Contains(err.Error(), "unsupported dialect") {
		t.Errorf("postgres dialect was not handled: %v", err)
	}
}
