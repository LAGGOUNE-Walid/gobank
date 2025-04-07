package storage

import (
	"database/sql"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

// SetupTestDB sets up a test SQLite in-memory database with migrations applied.
func SetupTestDB(t *testing.T) *SqliteStore {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal("failed to open database:", err)
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		t.Fatal("failed to create database instance:", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../migrations",
		"sqlite3", driver)
	if err != nil {
		t.Fatal("failed to initialize migrate:", err)
	}

	m.Down() // refresh
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatal("failed to apply migrations:", err)
	}

	store := &SqliteStore{db: db}

	return store
}
