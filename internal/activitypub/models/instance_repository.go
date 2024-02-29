package models

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/storage"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInstanceNotFound = errors.New("reading unknown instance from store")

type InstanceRepository struct {
	locker *sync.RWMutex
	config *instance.FederationConfig
	store  storage.Storage
}

func NewInstanceRepository(config *instance.FederationConfig, storage storage.Storage) *InstanceRepository {
	return &InstanceRepository{
		&sync.RWMutex{},
		config,
		storage,
	}
}

func (r *InstanceRepository) Upsert(ctx context.Context, instance *Instance) (*Instance, error) {
	r.locker.Lock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		defer r.locker.Unlock()
		cancel()
	}()

	result := tx.Clauses(clause.OnConflict{
		UpdateAll: false,
	}).Create(&instance)
	if result.Error != nil {
		return nil, fmt.Errorf("upsert instance for actor %s: %w", instance.Actor.ActorIri, result.Error)
	}

	return instance, nil
}

func (r *InstanceRepository) GetInstanceByActorIri(ctx context.Context, iri *url.URL) (*Instance, error) {
	r.locker.RLock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

	var shigInstance Instance

	result := tx.Preload("Actor").Table("actors").Select("*").Where("actor_iri=?", iri.String()).Joins("left join instances on instances.actor_id = actors.id").First(&shigInstance)
	if result.Error != nil {
		err := fmt.Errorf("finding instance by iri %s: %w", iri.String(), result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrInstanceNotFound)
		}
		return nil, err
	}

	return &shigInstance, nil

}
