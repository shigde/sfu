package store

import (
	"database/sql"
	"fmt"
)

type StorageConfig struct {
	Name       string `mapstructure:"name"`
	DataSource string `mapstructure:"dataSource"`
}
type Storage struct {
	db *sql.DB
}

func NewStorage(config StorageConfig) (*Storage, error) {
	if config.Name == "sqlite3" {
		return newSqlite(config.DataSource)
	}
	return nil, fmt.Errorf("creating storage %s not supported", config.Name)
}

func (s Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("closing database: %w", err)
	}
	return nil
}
