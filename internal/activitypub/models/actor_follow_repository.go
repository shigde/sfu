package models

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"gorm.io/gorm"
)

var ErrActorFollowNotFound = errors.New("actor follow not found")

type ActorFollowRepository struct {
	locker   *sync.RWMutex
	property *instance.Property
	storage  instance.Storage
}

func NewActorFollowRepository(property *instance.Property, storage instance.Storage) *ActorFollowRepository {
	return &ActorFollowRepository{
		&sync.RWMutex{},
		property,
		storage,
	}
}

func (r *ActorRepository) GetActorFollowForActorId(ctx context.Context, actorId uint) ([]*ActorFollow, error) {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

	var actorFollows []*ActorFollow
	results := tx.Where("actorId = ? AND state = ?", actorId, "accepted").Find(&actorFollows)
	if results.Error != nil {
		err := fmt.Errorf("finding actor follows for actor %d: %w", actorId, results.Error)
		if errors.Is(results.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrActorFollowNotFound)
		}
		return nil, err
	}

	return actorFollows, nil
}
