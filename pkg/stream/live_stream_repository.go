package stream

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrStreamNotFound = errors.New("reading unknown live stream from store")

type LiveStreamRepository struct {
	locker *sync.RWMutex
	store  storage
}

func NewLiveStreamRepository(store storage) (*LiveStreamRepository, error) {
	db := store.GetDatabase()
	if err := db.AutoMigrate(&LiveStream{}); err != nil {
		return nil, fmt.Errorf("migrating the space schema: %w", err)
	}
	return &LiveStreamRepository{
		&sync.RWMutex{},
		store,
	}, nil
}

func (r *LiveStreamRepository) Add(ctx context.Context, liveStream *LiveStream) (string, error) {
	r.locker.Lock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	if liveStream.Id == "" {
		liveStream.Id = uuid.New().String()
	}

	result := tx.Create(liveStream)
	if result.Error != nil || result.RowsAffected != 1 {
		return "", fmt.Errorf("adding live stream: %w", result.Error)
	}
	return liveStream.Id, nil
}

func (r *LiveStreamRepository) All(ctx context.Context) ([]LiveStream, error) {
	r.locker.RLock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

	var streams []LiveStream
	result := tx.Model(&LiveStream{}).Limit(501).Find(&streams)

	if result.Error != nil {
		return nil, fmt.Errorf("fetching all streams %w", result.Error)
	}

	return streams, nil
}

func (r *LiveStreamRepository) FindById(ctx context.Context, id string) (*LiveStream, error) {
	r.locker.RLock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

	liveStream := &LiveStream{Id: id}

	result := tx.First(liveStream)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("finding stream by id %s: %s: %w", id, ErrStreamNotFound, result.Error)
		}
		return nil, fmt.Errorf("finding stream by id %s: %w", id, result.Error)
	}

	return liveStream, nil
}

func (r *LiveStreamRepository) Delete(ctx context.Context, id string) error {
	r.locker.Lock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		defer r.locker.Unlock()
		cancel()
	}()

	result := tx.Delete(&LiveStream{Id: id})
	if result.Error != nil {
		return fmt.Errorf("deleting stream by id %s: %w", id, result.Error)
	}
	return nil
}

func (r *LiveStreamRepository) Contains(ctx context.Context, id string) bool {
	r.locker.RLock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

	var count int64
	tx.Find(&LiveStream{Id: id}).Count(&count)
	return count == 1
}

func (r *LiveStreamRepository) Update(ctx context.Context, stream *LiveStream) error {
	r.locker.Lock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		defer r.locker.Unlock()
		cancel()
	}()

	result := tx.Save(stream)
	if result.Error != nil {
		return fmt.Errorf("updating stream %s: %w", stream.Id, result.Error)
	}
	return nil
}

func (r *LiveStreamRepository) Len(ctx context.Context) int64 {
	r.locker.RLock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

	var count int64
	tx.Model(&LiveStream{}).Count(&count)
	return count
}

func (r *LiveStreamRepository) getStoreWithContext(ctx context.Context) (*gorm.DB, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeOut)
	db := r.store.GetDatabase()
	tx := db.WithContext(ctx)
	return tx, cancel
}
