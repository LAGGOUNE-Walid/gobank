package storage

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteStore struct {
	db *sql.DB
}

func NewSqliteStore() (*SqliteStore, error) {
	connection, err := sql.Open("sqlite3", "./db/database.sqlite")
	if err != nil {
		return nil, err
	}
	err = connection.Ping()
	if err != nil {
		log.Fatalf("Error connecting database: %v\n", err)
	}
	return &SqliteStore{db: connection}, nil

}
