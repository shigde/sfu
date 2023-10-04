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

func (r *FollowRepository) GetActorFollowByIri(ctx context.Context, iri string) (*ActorFollow, error) {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer cancel()

	actorFollow := &ActorFollow{Iri: iri}
	result := tx.First(actorFollow)
	if result.Error != nil {
		err := fmt.Errorf("finding actor follow for iri %s: %w", iri, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrActorNotFound)
		}
		return nil, err
	}

	return actorFollow, nil
}

func (r *FollowRepository) GetFollowByIri(ctx context.Context, iri string) (*Follow, error) {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer cancel()
	//Iri           string `gorm:"iri;not null;index"`
	//ActorId       uint   `gorm:"actor_id;not null"`
	//TargetActorId uint   `gorm:"target_actor_id;not null"`
	//State         string `gorm:"state;not null"`
	follow := &Follow{}
	result := tx.Table("actor_fallows").Select("actor_fallows.iri, actors.target_actor_id").Joins("left join emails on emails.user_id = users.id").Scan(&follow)

	if result.Error != nil {
		err := fmt.Errorf("finding actor follow for iri %s: %w", iri, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrActorNotFound)
		}
		return nil, err
	}

	return follow, nil
}

func (r *FollowRepository) GetActorFollowsFromActorId(ctx context.Context, actorId uint) ([]*ActorFollow, error) {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer cancel()

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

func (r *FollowRepository) GetActorFollowers(ctx context.Context, actorId uint) ([]*ActorFollow, error) {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer cancel()

	var actorFollows []*ActorFollow
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

func (r *FollowRepository) UpdateFollower(ctx context.Context, falower *ActorFollow) error {
	return nil
}
