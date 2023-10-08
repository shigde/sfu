package services

import (
	"context"
	"fmt"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/outbox"
	"github.com/shigde/sfu/internal/activitypub/remote"
)

type ActorService struct {
	config   *instance.FederationConfig
	actorRep *models.ActorRepository
	sender   *outbox.Sender
}

func NewActorService(config *instance.FederationConfig, actorRep *models.ActorRepository, sender *outbox.Sender) *ActorService {
	return &ActorService{
		config: config, actorRep: actorRep, sender: sender}
}

func (a *ActorService) GetLocalInstanceActor(ctx context.Context) (*models.Actor, error) {
	actor, err := a.actorRep.GetActorForUserName(ctx, a.config.InstanceUsername)
	if err != nil {
		return nil, fmt.Errorf("reading local instance actor from db: %w", err)
	}
	return actor, nil
}

func (a *ActorService) CreateActorFromRemoteAccount(ctx context.Context, accountIri string, localInstanceActor *models.Actor) (*models.Actor, error) {
	req, err := a.sender.GetSignedRequest(localInstanceActor.GetActorIri(), accountIri)
	if err != nil {
		return nil, fmt.Errorf("building signed actor request to fetch remote account: %w", err)
	}

	actor, err := remote.FetchAccountAsActor(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetching account from remote instance to build an actor: %w", err)
	}

	actor, err = a.actorRep.Upsert(ctx, actor)
	if err != nil {
		return nil, fmt.Errorf("insert actor in dab if not exists: %w", err)
	}
	return actor, nil
}
