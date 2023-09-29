package models

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"gorm.io/gorm"
)

type IriType int64

const (
	InboxIri IriType = iota
	ActorIri
)

var ErrActorNotFound = errors.New("actor not found")

type ActorRepository struct {
	locker   *sync.RWMutex
	property *instance.Property
	storage  instance.Storage
}

func NewActorRepository(property *instance.Property, storage instance.Storage) *ActorRepository {
	return &ActorRepository{
		&sync.RWMutex{},
		property,
		storage,
	}
}

func (r *ActorRepository) Add(ctx context.Context, actor *Actor) (*Actor, error) {
	return actor, nil
}

func (r *ActorRepository) GetActorForIRI(ctx context.Context, iri *url.URL, iriType IriType) (*Actor, error) {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer cancel()

	var actor *Actor
	switch iriType {
	case InboxIri:
		actor = &Actor{InboxIri: iri.String()}
	case ActorIri:
		actor = &Actor{InboxIri: iri.String()}
	default:
		return nil, errors.New(fmt.Sprintf("wrong iri type iriType: %d", iriType))
	}

	result := tx.First(actor)
	if result.Error != nil {
		err := fmt.Errorf("finding actor for IRI %s: %w", iri, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrActorNotFound)
		}
		return nil, err
	}

	return actor, nil
}

func (r *ActorRepository) GetAllActorsByIds(ctx context.Context, actorIds []uint) ([]*Actor, error) {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer cancel()

	var actors []*Actor
	results := tx.Where("id IN ?", actorIds).Find(&actors)
	if results.Error != nil {
		err := fmt.Errorf("finding actors by list: %w", results.Error)
		if errors.Is(results.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrActorNotFound)
		}
		return nil, err
	}

	return actors, nil

}
