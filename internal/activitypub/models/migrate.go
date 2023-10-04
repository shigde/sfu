package models

import (
	"errors"
	"fmt"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"golang.org/x/exp/slog"
	"gorm.io/gorm"
)

func Migrate(config *instance.FederationConfig, storage instance.Storage) error {
	storage.GetDatabase()
	db := storage.GetDatabase()

	if err := db.AutoMigrate(&ActorFollow{}, &Actor{}, &Server{}); err != nil {
		return fmt.Errorf("migrating the space schema: %w", err)
	}

	if db.Migrator().HasTable(&Actor{}) {
		if err := db.First(&Actor{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Info("creating instance actor")
			instanceActor, err := newInstanceActor(config.InstanceUrl, config.InstanceUsername)
			if err != nil {
				return fmt.Errorf("creating instance actor: %w", err)
			}
			db.Create(instanceActor)
		}
	}

	return nil
}
