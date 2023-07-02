package storage

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSqlite(dataSourceName string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connecting sqlite3 database: %w", err)
	}
	return &Store{db: db}, nil
}
