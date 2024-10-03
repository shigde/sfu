package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/storage"
	"gorm.io/gorm"
)

var ErrAccountNotFound = errors.New("account not found")

type AccountRepository struct {
	locker *sync.RWMutex
	store  storage.Storage
}

func NewAccountRepository(store storage.Storage) *AccountRepository {
	return &AccountRepository{
		&sync.RWMutex{},
		store,
	}
}

func (r *AccountRepository) findByUserName(ctx context.Context, user string) (*Account, error) {
	r.locker.RLock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

	var account Account

	result := tx.Where("user = ?", user).First(&account)
	if result.Error != nil {
		err := fmt.Errorf("finding stream by uuid %s: %w", user, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrAccountNotFound)
		}
		return nil, err
	}

	return &account, nil
}

func (r *AccountRepository) Add(ctx context.Context, account *Account) (string, error) {
	r.locker.Lock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	if len(account.UUID) == 0 {
		account.UUID = uuid.NewString()
	}

	result := tx.Create(account)
	if result.Error != nil || result.RowsAffected != 1 {
		return "", fmt.Errorf("adding account: %w", result.Error)
	}
	return account.UUID, nil
}

func (r *AccountRepository) AddVerificationToken(ctx context.Context, token *AccountVerificationToken) error {
	r.locker.Lock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	result := tx.Create(token)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("adding token: %w", result.Error)
	}
	return nil
}
