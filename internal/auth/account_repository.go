package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/shigde/sfu/internal/storage"
	"gorm.io/gorm"
)

var ErrAccountNotFound = errors.New("account not found")

type accountRepository struct {
	locker *sync.RWMutex
	store  *storage.Store
}

func newAccountRepository(store *storage.Store) (*accountRepository, error) {
	return &accountRepository{
		&sync.RWMutex{},
		store,
	}, nil
}

func (r accountRepository) findByUserName(ctx context.Context, user string) (*Account, error) {
	r.locker.RLock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

	account := &Account{User: user}

	result := tx.First(account)
	if result.Error != nil {
		err := fmt.Errorf("finding stream by uuid %s: %w", user, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrAccountNotFound)
		}
		return nil, err
	}

	return account, nil
}
