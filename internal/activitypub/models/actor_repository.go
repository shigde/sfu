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

type IriType int64

const (
	InboxIri IriType = iota
	ActorIri
)

var ErrActorNotFound = errors.New("actor not found")

type ActorRepository struct {
	locker *sync.RWMutex
	config *instance.FederationConfig
	store  storage.Storage
}

func NewActorRepository(config *instance.FederationConfig, storage storage.Storage) *ActorRepository {
	return &ActorRepository{
		&sync.RWMutex{},
		config,
		storage,
	}
}

func (r *ActorRepository) Add(_ context.Context, actor *Actor) (*Actor, error) {
	return actor, nil
}

func (r *ActorRepository) Upsert(ctx context.Context, actor *Actor) (*Actor, error) {
	r.locker.Lock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		defer r.locker.Unlock()
		cancel()
	}()

	result := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "actor_iri"}},
		UpdateAll: true,
	}).Create(&actor)
	if result.Error != nil {
		return nil, fmt.Errorf("creating actor for name %s: %w", actor.PreferredUsername, result.Error)
	}

	return actor, nil
}

func (r *ActorRepository) GetActorForUserName(ctx context.Context, name string) (*Actor, error) {
	r.locker.RLock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

	actor := &Actor{PreferredUsername: name}
	result := tx.First(actor)
	if result.Error != nil {
		err := fmt.Errorf("finding actor for name %s: %w", name, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrActorNotFound)
		}
		return nil, err
	}

	return actor, nil
}

func (r *ActorRepository) GetActorByActorIRI(ctx context.Context, actorIri *url.URL) (*Actor, error) {
	r.locker.RLock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()
	var actor Actor

	result := tx.Where("actor_iri=?", actorIri.String()).First(&actor)
	if result.Error != nil {
		err := fmt.Errorf("finding actor for IRI %s: %w", actorIri.String(), result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, ErrActorNotFound)
		}
		return nil, err
	}
	return &actor, nil
}

// GetActorForIRI
// @deprecated
func (r *ActorRepository) GetActorForIRI(ctx context.Context, iri *url.URL, iriType IriType) (*Actor, error) {
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
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
	r.locker.RLock()
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer func() {
		defer r.locker.RUnlock()
		cancel()
	}()

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

func (r *ActorRepository) GetPublicKey(ctx context.Context, actorIRI *url.URL) (string, error) {
	actor, err := r.GetActorByActorIRI(ctx, actorIRI)
	if err != nil {
		return "", fmt.Errorf("getting public key: %w", err)
	}
	return actor.PublicKey, nil
}

func (r *ActorRepository) GetPrivateKey(ctx context.Context, actorIRI *url.URL) (string, error) {
	actor, err := r.GetActorByActorIRI(ctx, actorIRI)
	if err != nil {
		return "", fmt.Errorf("getting private key: %w", err)
	}
	return actor.PrivateKey.String, nil
}

func (r *ActorRepository) GetKeyPair(ctx context.Context, actorIRI *url.URL) (publicKey string, privateKey string, err error) {
	actor, err := r.GetActorByActorIRI(ctx, actorIRI)
	if err != nil {
		return "", "", fmt.Errorf("getting key pair: %w", err)
	}
	return actor.PublicKey, actor.PrivateKey.String, nil
}
