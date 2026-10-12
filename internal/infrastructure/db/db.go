package db

import (
	"fmt"
	"os"
	"path/filepath"

	"labshop/internal/adapter/repository/sqlite"
)

// Init ensures the database directory exists and opens the SQLite connection pool with migrations applied.
func Init(dbPath string) (*sqlite.DB, error) {
	if dbPath != ":memory:" && !filepath.IsAbs(dbPath) {
		dir := filepath.Dir(dbPath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create database directory %s: %w", dir, err)
			}
		}
	} else if filepath.IsAbs(dbPath) {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory %s: %w", dir, err)
		}
	}

	database, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database at %s: %w", dbPath, err)
	}

	return database, nil
}
