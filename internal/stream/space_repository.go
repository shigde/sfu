package stream

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrSpaceNotFound = errors.New("reading unknown space in store")

type SpaceRepository struct {
	locker *sync.RWMutex
	store  storage
}

func NewSpaceRepository(store storage) *SpaceRepository {
	return &SpaceRepository{
		&sync.RWMutex{},
		store,
	}
}

func (r *SpaceRepository) Add(ctx context.Context, space *Space) (string, error) {
	r.locker.Lock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	if len(space.Identifier) == 0 {
		space.Identifier = uuid.NewString()
	}

	result := tx.Create(space)
	if result.Error != nil || result.RowsAffected != 1 {
		return "", fmt.Errorf("adding live stream: %w", result.Error)
	}
	return space.Identifier, nil
}

func (r *SpaceRepository) GetByIdentifier(ctx context.Context, identifier string) (*Space, error) {
	r.locker.Lock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	var space Space
	result := tx.Where("identifier = ?", identifier).First(&space)
	if result.Error != nil {
		err := fmt.Errorf("finding space by space identifier %s: %w", identifier, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrSpaceNotFound)
		}
		return nil, err
	}

	return &space, nil
}

func (r *SpaceRepository) CreateWithIdentifier(ctx context.Context, identifier string) (*Space, error) {
	r.locker.Lock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	space := &Space{Identifier: identifier}
	result := tx.Create(space)
	if result.Error != nil {
		return nil, fmt.Errorf("creating space for name %s: %w", space.Identifier, result.Error)
	}

	return space, nil
}

func (r *SpaceRepository) Delete(ctx context.Context, identifier string) error {
	r.locker.Lock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	result := tx.Where("identifier = ?", identifier).Delete(&Space{})
	if result.Error != nil {
		return fmt.Errorf("deleting space by space id %s: %w", identifier, result.Error)
	}
	return nil
}

func (r *SpaceRepository) Len(ctx context.Context) int64 {
	r.locker.RLock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		r.locker.RUnlock()
		cancel()
	}()

	var count int64
	tx.Model(&Space{}).Count(&count)
	return count
}

func (r *SpaceRepository) getStoreWithContext(ctx context.Context) (*gorm.DB, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeOut)
	db := r.store.GetDatabase()
	tx := db.WithContext(ctx)
	return tx, cancel
}
