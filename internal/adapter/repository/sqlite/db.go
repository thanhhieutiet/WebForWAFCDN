package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type txKey struct{}

func withTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func fromContext(ctx context.Context) *sql.Tx {
	tx, _ := ctx.Value(txKey{}).(*sql.Tx)
	return tx
}

type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type DB struct {
	db *sql.DB
}

func Open(dsn string) (*DB, error) {
	connStr := dsn
	if !strings.Contains(connStr, "_pragma") {
		separator := "?"
		if strings.Contains(connStr, "?") {
			separator = "&"
		}
		connStr = fmt.Sprintf("%s%s_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)", connStr, separator)
	}

	sqlDB, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	sqlDB.SetMaxOpenConns(1) // SQLite works best with 1 writer connection
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(0)

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to ping sqlite db: %w", err)
	}

	database := &DB{db: sqlDB}
	if err := database.Migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return database, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) SQLDB() *sql.DB {
	return d.db
}

func (d *DB) exec(ctx context.Context) executor {
	if tx := fromContext(ctx); tx != nil {
		return tx
	}
	return d.db
}

// Migrate executes all pending SQL migration files.
func (d *DB) Migrate() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations dir: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		var count int
		err := d.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", file).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to query migration status for %s: %w", file, err)
		}
		if count > 0 {
			continue // Already applied
		}

		content, err := migrationsFS.ReadFile("migrations/" + file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		tx, err := d.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", file, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		_, err = tx.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", file, time.Now())
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", file, err)
		}
	}

	return nil
}

type TxManager struct {
	db *DB
}

func NewTxManager(db *DB) *TxManager {
	return &TxManager{db: db}
}

func (tm *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if fromContext(ctx) != nil {
		return fn(ctx)
	}

	tx, err := tm.db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := withTx(ctx, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
