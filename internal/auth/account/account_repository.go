package account

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

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

func (r *AccountRepository) findActiveByEmail(ctx context.Context, email string) (*Account, error) {
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

func (r *AccountRepository) findByEmail(ctx context.Context, email string) (*Account, error) {
	r.locker.RLock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

	var account Account

	result := tx.Preload("Actor").Where("email = ?", email).First(&account)
	if result.Error != nil {
		err := fmt.Errorf("finding all account by email %s: %w", email, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrAccountNotFound)
		}
		return nil, err
	}

	return &account, nil
}

func (r *AccountRepository) findActiveByUuid(ctx context.Context, userUuid *uuid.UUID) (*Account, error) {
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

func (r *AccountRepository) deleteByUuid(ctx context.Context, userUuid string) error {
	r.locker.Lock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		defer r.locker.Unlock()
		cancel()
	}()

	var account Account

	result := tx.Unscoped().Where("uuid = ?", userUuid).Delete(&account)
	if result.Error != nil {
		return fmt.Errorf("delete account by uuid %s: %w", userUuid, result.Error)
	}

	return nil
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

	lastHour := time.Now().Add(-time.Hour)
	result := tx.Preload("Account").Where("uuid = ? AND verified = ? AND created_at > ?", token, false, lastHour).First(&verificationToken)
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
		return fmt.Errorf("deactivate token: %s: %w", token, result.Error)
	}

	saved = tx.Save(verificationToken.Account)
	if saved.Error != nil {
		return fmt.Errorf("activating account by verification token %s: %w", token, result.Error)
	}

	return nil
}

func (r *AccountRepository) update(ctx context.Context, account *Account) error {
	r.locker.Lock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	saved := tx.Save(&account)
	if saved.Error != nil {
		return fmt.Errorf("update account %s: %w", account.ID, saved.Error)
	}

	return nil
}

func (r *AccountRepository) RedeemPassForgetToken(ctx context.Context, token string) (*VerificationToken, error) {
	r.locker.Lock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	var verificationToken VerificationToken

	lastHour := time.Now().Add(-time.Hour)
	result := tx.Preload("Account").Where("uuid = ? AND verified = ? AND created_at > ?", token, false, lastHour).First(&verificationToken)
	if result.Error != nil {
		err := fmt.Errorf("seraching redeem pass forget token %s: %w", token, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrTokenNotFound)
		}
		return nil, err
	}

	verificationToken.Verified = true
	saved := tx.Save(&verificationToken)
	if saved.Error != nil {
		return nil, fmt.Errorf("redeem pass forget token %s: %w", token, result.Error)
	}

	return &verificationToken, nil
}
