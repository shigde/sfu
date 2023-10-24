package models

import (
	"context"
	"fmt"
	"sync"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/storage"
	"gorm.io/gorm/clause"
)

type VideoRepository struct {
	locker  *sync.RWMutex
	config  *instance.FederationConfig
	storage storage.Storage
}

func NewVideoRepository(config *instance.FederationConfig, storage storage.Storage) *VideoRepository {
	return &VideoRepository{
		&sync.RWMutex{},
		config,
		storage,
	}
}

func (r *VideoRepository) Upsert(ctx context.Context, video *Video) (*Video, error) {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer cancel()

	result := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "iri"}},
		UpdateAll: true,
	}).Create(&video)
	if result.Error != nil {
		return nil, fmt.Errorf("upsert video for id %d: %w", video.ID, result.Error)
	}

	return video, nil
}

func (r *VideoRepository) DeleteByIri(ctx context.Context, iri string) error {
	tx, cancel := r.storage.GetDatabaseWithContext(ctx)
	defer cancel()

	result := tx.Unscoped().Delete(&Video{}, "iri = ?", iri)
	if result.Error != nil {
		return fmt.Errorf("delete video for iri %s: %w", iri, result.Error)
	}
	return nil
}
