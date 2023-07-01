package store

import (
	"database/sql"
	"fmt"
)

func newSqlite(dataSourceName string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("connecting sqlite3 database: %w", err)
	}
	return &Storage{db: db}, nil
}
