package services

import (
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
)

type VideoService struct {
	config       *instance.FederationConfig
	actorService *ActorService
	videoRep     *models.VideoRepository
}

func NewVideoService(config *instance.FederationConfig, actorService *ActorService, videoRep *models.VideoRepository) *VideoService {
	return &VideoService{config: config, actorService: actorService, videoRep: videoRep}
}
