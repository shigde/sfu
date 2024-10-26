package account

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
var ErrTokenNotFound = errors.New("token not found")

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

	result := tx.Where("user = ? AND active = ?", user, true).First(&account)
	if result.Error != nil {
		err := fmt.Errorf("finding account by name %s: %w", user, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrAccountNotFound)
		}
		return nil, err
	}

	return &account, nil
}

func (r *AccountRepository) findByEmail(ctx context.Context, email string) (*Account, error) {
	r.locker.RLock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

	var account Account

	result := tx.Where("email = ? AND active = ?", email, true).First(&account)
	if result.Error != nil {
		err := fmt.Errorf("finding account by email %s: %w", email, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrAccountNotFound)
		}
		return nil, err
	}

	return &account, nil
}

func (r *AccountRepository) findByUuid(ctx context.Context, userUuid *uuid.UUID) (*Account, error) {
	r.locker.RLock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

	var account Account

	result := tx.Where("uuid = ? AND active = ?", userUuid.String(), true).First(&account)
	if result.Error != nil {
		err := fmt.Errorf("finding account by uuid %s: %w", userUuid.String(), result.Error)
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

func (r *AccountRepository) AddVerificationToken(ctx context.Context, token *VerificationToken) error {
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

func (r *AccountRepository) RedeemAccountVerificationToken(ctx context.Context, token string) error {
	r.locker.Lock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	var verificationToken VerificationToken

	result := tx.Preload("Account").Where("token = ? AND verified = ?", token, false).First(&verificationToken)
	if result.Error != nil {
		err := fmt.Errorf("seraching redeem account verification token %s: %w", token, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.Join(err, ErrTokenNotFound)
		}
		return err
	}

	verificationToken.Verified = true
	verificationToken.Account.Active = true
	saved := tx.Save(&verificationToken)
	if saved.Error != nil {
		return fmt.Errorf("activating account by verification token %s: %w", token, result.Error)
	}

	return nil
}
