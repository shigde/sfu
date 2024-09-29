package migration

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth"
	"gorm.io/gorm"
)

func buildInstance(db *gorm.DB, instanceUrl *url.URL, instanceName string) (*models.Instance, error) {

	actor, err := models.NewInstanceActor(instanceUrl, instanceName)
	if err != nil {
		return nil, fmt.Errorf("migration building instance actor: %w", err)
	}

	var currentActor models.Actor
	result := db.Where("actor_iri=?", actor.ActorIri).First(&currentActor)

	// somthing get wrong
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("migration loading actor: %w", result.Error)
	}

	// load exiting instance because actor exists
	if result.Error == nil {
		var currenInstance models.Instance
		result = db.Preload("Actor").Where("actor_id=?", currentActor.ID).First(&currenInstance)
		if result.Error != nil {
			return nil, fmt.Errorf("migration loading existing instance: %w", result.Error)
		}
		return &currenInstance, nil
	}

	// create instance
	db.Save(actor)
	userId := creatUserId(instanceName, instanceUrl)
	account := auth.CreateInstanceAccount(userId, actor)
	db.Create(account)

	serverInstance := models.NewInstance(actor)
	db.Save(serverInstance)
	return serverInstance, nil
}
