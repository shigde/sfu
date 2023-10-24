package storage

import (
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
