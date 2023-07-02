package stream

import (
	"context"
	"fmt"
	"sync"

	"gorm.io/gorm"
)

type SpaceRepository struct {
	locker *sync.RWMutex
	space  map[string]*Space
	db     *gorm.DB
	lobby  lobbyAccessor
}

func newSpaceRepository(lobby lobbyAccessor, store storage) (*SpaceRepository, error) {
	space := make(map[string]*Space)
	db := store.GetDatabase()
	if err := db.AutoMigrate(&Space{}); err != nil {
		return nil, fmt.Errorf("migrating the space schema: %w", err)
	}

	return &SpaceRepository{
		&sync.RWMutex{},
		space,
		db,
		lobby,
	}, nil
}

func (r *SpaceRepository) GetOrCreateSpace(ctx context.Context, id string) *Space {
	r.locker.Lock()
	ctx, cancel := context.WithTimeout(ctx, queryTimeOut)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()
	space := newSpace(id, r.lobby)
	tx := r.db.WithContext(ctx)
	result := tx.FirstOrCreate(&space)
	if result.Error != nil || result.RowsAffected != 1 {
		return nil
	}
	return space
}

func (r *SpaceRepository) GetSpace(ctx context.Context, id string) (*Space, bool) {
	r.locker.Lock()
	ctx, cancel := context.WithTimeout(ctx, queryTimeOut)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()

	space := newSpace(id, r.lobby)
	tx := r.db.WithContext(ctx)
	result := tx.Find(&space)
	if result.Error != nil || result.RowsAffected != 1 {
		return nil, false
	}
	return space, true
}

func (r *SpaceRepository) Delete(ctx context.Context, id string) bool {
	r.locker.Lock()
	ctx, cancel := context.WithTimeout(ctx, queryTimeOut)
	defer func() {
		r.locker.Unlock()
		cancel()
	}()
	if _, ok := r.space[id]; ok {
		delete(r.space, id)
		return true
	}

	space := newSpace(id, r.lobby)
	tx := r.db.WithContext(ctx)
	result := tx.Delete(space)
	if result.Error != nil || result.RowsAffected != 1 {
		return false
	}
	return true
}

func (r *SpaceRepository) Len(ctx context.Context) int64 {
	r.locker.RLock()
	ctx, cancel := context.WithTimeout(ctx, queryTimeOut)
	defer func() {
		r.locker.RUnlock()
		cancel()
	}()
	tx := r.db.WithContext(ctx)
	var count int64
	tx.Model(&Space{}).Count(&count)
	return count
}
