package auth

import (
	"fmt"
	"sync"

	"github.com/shigde/sfu/internal/storage"
)

type accountRepository struct {
	locker *sync.RWMutex
	store  *storage.Store
}

func newAccountRepository(store *storage.Store) (*accountRepository, error) {
	db := store.GetDatabase()
	if err := db.AutoMigrate(&Account{}); err != nil {
		return nil, fmt.Errorf("migrating the space schema: %w", err)
	}
	return &accountRepository{
		&sync.RWMutex{},
		store,
	}, nil
}
