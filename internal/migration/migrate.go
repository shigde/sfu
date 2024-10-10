package migration

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/lobby"
	"github.com/shigde/sfu/internal/storage"
	"github.com/shigde/sfu/internal/stream"
	"golang.org/x/exp/slog"
	"gorm.io/gorm"
)

func Migrate(config *instance.FederationConfig, storage storage.Storage) error {
	storage.GetDatabase()
	db := storage.GetDatabase()

	if err := db.AutoMigrate(
		&models.Video{},
		&models.Follow{},
		&models.Actor{},
		&models.Instance{},
		&account.Account{},
		&account.AccountVerificationToken{},
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
			shigInstance := models.NewInstance(instanceActor)
			db.Create(shigInstance)

			for _, inst := range config.TrustedInstances {
				if err := buildTrustedInstanceAccount(db, inst); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func buildTrustedInstanceAccount(db *gorm.DB, trustedInstance instance.TrustedInstance) error {
	actorIri, err := url.Parse(trustedInstance.Actor)
	if err != nil {
		return fmt.Errorf("migration parsing trusted instance actor id: %w", err)
	}
	actorId := fmt.Sprintf("%s@%s", trustedInstance.Name, actorIri.Host)
	actor, err := models.NewTrustedInstanceActor(actorIri, trustedInstance.Name)
	if err != nil {
		return fmt.Errorf("migration building actor: %w", err)
	}

	db.Create(actor)
	var savedActor models.Actor

	result := db.Where("actor_iri=?", actorIri.String()).First(&savedActor)
	if result.Error != nil {
		return fmt.Errorf("migration loading actor: %w", result.Error)
	}

	account := account.CreateInstanceAccount(actorId, &savedActor)
	db.Create(account)

	trInstance := models.NewInstance(&savedActor)
	db.Create(trInstance)
	return nil
}
