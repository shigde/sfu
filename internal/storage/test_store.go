package storage

import (
	"context"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestStore struct {
	db *gorm.DB
}

func NewTestStore() *TestStore {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	return &TestStore{db}
}

func (s *TestStore) GetDatabase() *gorm.DB {
	return s.db
}

func (s *TestStore) GetDatabaseWithContext(ctx context.Context) (*gorm.DB, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeOut)
	tx := s.db.WithContext(ctx)
	return tx, cancel
}
