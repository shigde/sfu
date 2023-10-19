package stream

import (
	"context"
	"errors"
	"fmt"
	"sync"

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

func (r *SpaceRepository) GetSpace(ctx context.Context, id string) (*Space, error) {
	r.locker.Lock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	space, err := newSpace(id)
	if err != nil {
		return nil, fmt.Errorf("finding space by space id %s: %w", id, err)
	}

	result := tx.First(space)
	if result.Error != nil {
		err = fmt.Errorf("finding space by space id %s: %w", id, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrSpaceNotFound)
		}
		return nil, err
	}

	return space, nil
}

func (r *SpaceRepository) GetSpaceByIdentifier(ctx context.Context, id string) (*Space, error) {
	r.locker.Lock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	space, err := newSpace(id)
	if err != nil {
		return nil, fmt.Errorf("finding space by space id %s: %w", id, err)
	}

	result := tx.First(space)
	if result.Error != nil {
		err = fmt.Errorf("finding space by space id %s: %w", id, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrSpaceNotFound)
		}
		return nil, err
	}

	return space, nil
}

func (r *SpaceRepository) Delete(ctx context.Context, id string) error {
	r.locker.Lock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	space, err := newSpace(id)
	if err != nil {
		return fmt.Errorf("deleting space by space id %s: %w", id, err)
	}

	result := tx.Delete(space)
	if result.Error != nil {
		return fmt.Errorf("deleting space by space id %s: %w", id, result.Error)
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
