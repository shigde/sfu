package actor

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/shigde/sfu/internal/activitypub"
	"gorm.io/gorm"
)

type Repository struct {
	locker  *sync.RWMutex
	storage activitypub.Storage
}

func NewRepository(storage activitypub.Storage) (*Repository, error) {
	storage.GetDatabase()
	db := storage.GetDatabase()

	if err := db.AutoMigrate(&ActivityPubActor{}); err != nil {
		return nil, fmt.Errorf("migrating the space schema: %w", err)
	}

	if db.Migrator().HasTable(&ActivityPubActor{}) {
		if err := db.First(&ActivityPubActor{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			//Insert seed data
		}
	}

	return &Repository{
		&sync.RWMutex{},
		storage,
	}, nil
}

func (r *Repository) Add(ctx context.Context, actor *ActivityPubActor) (*ActivityPubActor, error) {

	return actor, nil
}
