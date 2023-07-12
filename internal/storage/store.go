package storage

import (
	"fmt"

	"gorm.io/gorm"
)

type StorageConfig struct {
	Name       string `mapstructure:"name"`
	DataSource string `mapstructure:"dataSource"`
}
type Store struct {
	db *gorm.DB
}

func NewStore(config *StorageConfig) (*Store, error) {
	if config.Name == "sqlite3" {
		return newSqlite(config.DataSource)
	}
	return nil, fmt.Errorf("creating storage %s not supported", config.Name)
}

func (s *Store) GetDatabase() *gorm.DB {
	return s.db
}
