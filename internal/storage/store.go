package storage

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type StorageConfig struct {
	Name       string `mapstructure:"name"`
	DataSource string `mapstructure:"dataSource"`
}
type Store struct {
	db *gorm.DB
}

const queryTimeOut = 5 * time.Second

func NewStore(config *StorageConfig) (*Store, error) {
	if config.Name == "sqlite3" {
		return newSqlite(config.DataSource)
	}
	return nil, fmt.Errorf("creating storage %s not supported", config.Name)
}

func (s *Store) GetDatabase() *gorm.DB {
	return s.db
}

func (s *Store) GetDatabaseWithContext(ctx context.Context) (*gorm.DB, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeOut)
	tx := s.db.WithContext(ctx)
	return tx, cancel
}
