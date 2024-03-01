package services

import (
	"context"
	"net/url"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
)

type InstanceService struct {
	config       *instance.FederationConfig
	instanceRepo *models.InstanceRepository
}

func NewInstanceService(config *instance.FederationConfig, instanceRepo *models.InstanceRepository) *InstanceService {
	return &InstanceService{config: config, instanceRepo: instanceRepo}
}

func (s *InstanceService) getInstanceByActorIri(ctx context.Context, iri *url.URL) (*models.Instance, bool) {
	if shigInstance, err := s.instanceRepo.GetInstanceByActorIri(ctx, iri); err == nil {
		return shigInstance, true
	}
	return nil, false
}

func (s *InstanceService) upsertInstanceByActor(ctx context.Context, actor *models.Actor) (*models.Instance, error) {
	shigInstance := models.NewInstance(actor)
	return s.instanceRepo.Upsert(ctx, shigInstance)
}
