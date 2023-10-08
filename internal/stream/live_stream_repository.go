package stream

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/auth"
	"gorm.io/gorm"
)

var ErrStreamNotFound = errors.New("reading unknown live stream from store")

type LiveStreamRepository struct {
	locker *sync.RWMutex
	store  storage
}

func NewLiveStreamRepository(store storage) *LiveStreamRepository {
	return &LiveStreamRepository{
		&sync.RWMutex{},
		store,
	}
}

func (r *LiveStreamRepository) Add(ctx context.Context, liveStream *LiveStream) (string, error) {
	r.locker.Lock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	if len(liveStream.UUID.String()) == 0 {
		liveStream.UUID = uuid.New()
	}

	if liveStream.UUID == uuid.Nil {
		liveStream.UUID = uuid.New()
	}

	result := tx.Create(liveStream)
	if result.Error != nil || result.RowsAffected != 1 {
		return "", fmt.Errorf("adding live stream: %w", result.Error)
	}
	return liveStream.UUID.String(), nil
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

func (r *LiveStreamRepository) FindByUuid(ctx context.Context, streamUUID string) (*LiveStream, error) {
	r.locker.RLock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()
	UUID, err := uuid.Parse(streamUUID)
	if err != nil {
		return nil, fmt.Errorf("parsing UUID: %w", err)
	}

	liveStream := &LiveStream{UUID: UUID}

	result := tx.First(liveStream)
	if result.Error != nil {
		err := fmt.Errorf("finding stream by uuid %s: %w", streamUUID, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrStreamNotFound)
		}
		return nil, err
	}

	return liveStream, nil
}

func (r *LiveStreamRepository) Delete(ctx context.Context, streamUUID string) error {
	r.locker.Lock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		defer r.locker.Unlock()
		cancel()
	}()
	UUID, err := uuid.Parse(streamUUID)
	if err != nil {
		return fmt.Errorf("parsing UUID: %w", err)
	}
	result := tx.Delete(&LiveStream{UUID: UUID})
	if result.Error != nil {
		return fmt.Errorf("deleting stream by id %s: %w", streamUUID, result.Error)
	}
	return nil
}

func (r *LiveStreamRepository) Contains(ctx context.Context, streamUUID string) bool {
	r.locker.RLock()
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

	UUID, err := uuid.Parse(streamUUID)
	if err != nil {
		return false
	}

	var count int64
	tx.Find(&LiveStream{UUID: UUID}).Count(&count)
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
		return fmt.Errorf("updating stream %s: %w", stream.UUID, result.Error)
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

func (r *LiveStreamRepository) UpsertLiveStream(ctx context.Context, stream *LiveStream) error {
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		cancel()
	}()

	account := auth.Account{User: stream.User}
	resultAc := tx.First(&account)
	if resultAc.Error != nil && !errors.Is(resultAc.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("serching account: %w", resultAc.Error)
	}

	space := Space{Identifier: stream.Space.Identifier}
	resultSp := tx.First(&space)
	if resultSp.Error != nil && !errors.Is(resultSp.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("serching space: %w", resultAc.Error)
	}

	if resultAc.RowsAffected > 0 {
		stream.Account = &account
	}

	if resultSp.RowsAffected > 0 {
		stream.Space = &space
	}

	result := tx.Create(stream)
	if result.Error != nil || result.RowsAffected == 0 {
		return fmt.Errorf("adding live stream: %w", result.Error)
	}
	return nil
}

func (r *LiveStreamRepository) DeleteByUuid(ctx context.Context, streamUUID string) error {
	tx, cancel := r.getStoreWithContext(ctx)
	defer func() {
		cancel()
	}()

	result := tx.Unscoped().Delete(&LiveStream{}, "uuid = ?", streamUUID)
	if result.Error != nil {
		return fmt.Errorf("deleting stream by uuid %s: %w", streamUUID, result.Error)
	}
	return nil

}
