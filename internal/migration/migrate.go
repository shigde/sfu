package migration

import (
	"errors"
	"fmt"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/lobby"
	"github.com/shigde/sfu/internal/stream"
	"golang.org/x/exp/slog"
	"gorm.io/gorm"
)

func Migrate(config *instance.FederationConfig, storage instance.Storage) error {
	storage.GetDatabase()
	db := storage.GetDatabase()

	if err := db.AutoMigrate(
		&models.Video{},
		&models.Follow{},
		&models.Actor{},
		&models.Server{},
		&auth.Account{},
		&lobby.LobbyEntity{},
		&stream.Space{},
		&stream.LiveStream{},
	); err != nil {
		return fmt.Errorf("migrating the space schema: %w", err)
	}

	if db.Migrator().HasTable(&models.Actor{}) {
		if err := db.First(&models.Actor{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Info("creating instance actor")
			instanceActor, err := models.NewInstanceActor(config.InstanceUrl, config.InstanceUsername)
			if err != nil {
				return fmt.Errorf("creating instance actor: %w", err)
			}
			db.Create(instanceActor)
		}
	}

	return nil
}
