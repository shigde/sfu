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

type FollowRepository struct {
	locker  *sync.RWMutex
	config  *instance.FederationConfig
	storage instance.Storage
}

func NewFollowRepository(config *instance.FederationConfig, storage instance.Storage) *FollowRepository {
	return &FollowRepository{
		&sync.RWMutex{},
		config,
		storage,
	}
}

func (r *FollowRepository) Add(ctx context.Context, follow *Follow) (*Follow, error) {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer cancel()

	results := tx.Create(follow)
	if results.Error != nil {
		return nil, fmt.Errorf("adding new actor follows for actor %d: %w", follow.ActorId, results.Error)
	}

	return follow, nil
}

func (r *FollowRepository) Update(ctx context.Context, follow *Follow) (*Follow, error) {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer cancel()

	results := tx.Save(follow)
	if results.Error != nil {
		return nil, fmt.Errorf("update actor follows for actor %d: %w", follow.ActorId, results.Error)
	}

	return follow, nil
}

func (r *FollowRepository) GetFollowByIri(ctx context.Context, iri string) (*Follow, error) {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer cancel()

	actorFollow := &Follow{Iri: iri}
	result := tx.Preload("Actor").Preload("TargetActor").First(actorFollow)
	if result.Error != nil {
		err := fmt.Errorf("finding actor follow for iri %s: %w", iri, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrActorNotFound)
		}
		return nil, err
	}

	return actorFollow, nil
}

func (r *FollowRepository) GetActorFollowsFromActorId(ctx context.Context, actorId uint) ([]*Follow, error) {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer cancel()

	var actorFollows []*Follow
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

func (r *FollowRepository) GetActorFollowers(ctx context.Context, actorId uint) ([]*Follow, error) {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer cancel()

	var actorFollows []*Follow
	results := tx.Where("targetActorId = ? AND state = ?", actorId, "accepted").Find(&actorFollows)
	if results.Error != nil {
		err := fmt.Errorf("finding follower for actor %d: %w", actorId, results.Error)
		if errors.Is(results.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrActorFollowNotFound)
		}
		return nil, err
	}

	return actorFollows, nil
}

func (r *FollowRepository) UpdateFollower(ctx context.Context, falower *Follow) error {
	return nil
}
