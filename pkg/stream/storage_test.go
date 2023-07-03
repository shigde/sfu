package stream

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testStore struct {
	db *gorm.DB
}

func newTestStore() *testStore {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	return &testStore{db}
}

func (s *testStore) GetDatabase() *gorm.DB {
	return s.db
}
