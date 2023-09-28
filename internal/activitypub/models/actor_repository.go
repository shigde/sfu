package models

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"golang.org/x/exp/slog"
	"gorm.io/gorm"
)

type ActorRepository struct {
	locker   *sync.RWMutex
	property *instance.Property
	storage  instance.Storage
}

func NewActorRepository(property *instance.Property, storage instance.Storage) (*ActorRepository, error) {
	storage.GetDatabase()
	db := storage.GetDatabase()

	if err := db.AutoMigrate(&Actor{}); err != nil {
		return nil, fmt.Errorf("migrating the space schema: %w", err)
	}

	if db.Migrator().HasTable(&Actor{}) {
		if err := db.First(&Actor{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Info("creating instance actor")
			instanceActor, err := newInstanceActor(property.InstanceUrl, "shig")
			if err != nil {
				return nil, fmt.Errorf("creating instance actor: %w", err)
			}
			db.Create(instanceActor)
		}
	}

	return &ActorRepository{
		&sync.RWMutex{},
		property,
		storage,
	}, nil
}

func (r *ActorRepository) Add(ctx context.Context, actor *Actor) (*Actor, error) {
	return actor, nil
}
