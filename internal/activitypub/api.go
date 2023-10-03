package activitypub

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/activitypub/crypto"
	"github.com/shigde/sfu/internal/activitypub/inbox"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/remote"
	"github.com/shigde/sfu/internal/activitypub/workerpool"
	"github.com/superseriousbusiness/activity/pub"
)

type ApApi struct {
	config           *instance.FederationConfig
	InstanceProperty *instance.Property
	Storage          instance.Storage
	actorRepo        *models.ActorRepository
	actor            pub.FederatingActor
	signer           *crypto.Signer
}

func NewApApi(config *instance.FederationConfig, storage instance.Storage) (*ApApi, error) {
	instanceProperty := instance.NewProperty(config)

	if err := models.Migrate(instanceProperty, storage); err != nil {
		return nil, fmt.Errorf("creation schema for federation: %w", err)
	}
	actorRepo := models.NewActorRepository(instanceProperty, storage)
	actorFollowRepo := models.NewActorFollowRepository(instanceProperty, storage)

	behavior := NewCommonBehavior()
	protocol := NewFederatingProtocol()
	database := NewDatabase(actorRepo, actorFollowRepo)
	clock := NewClock()

	actor := pub.NewFederatingActor(behavior, protocol, database, clock)

	signer := crypto.NewSigner(actorRepo)

	return &ApApi{
		config:           config,
		InstanceProperty: instanceProperty,
		Storage:          storage,
		actorRepo:        actorRepo,
		actor:            actor,
		signer:           signer,
	}, nil

}

func (a *ApApi) BoostrapApi(router *mux.Router) error {
	if err := extendRouter(router, a.config, a.actorRepo, a.signer); err != nil {
		return fmt.Errorf("extending router with federation endpoints: %w", err)
	}

	if a.config.Enable {
		resolver := remote.NewResolver(a.config, a.signer)
		workerpool.InitOutboundWorkerPool()
		inbox.InitInboxWorkerPool(resolver)
	}

	return nil
}

//// SendLive will send a "Go Live" message to followers.
//func SendLive() error {
//	return outbox.SendLive()
//}
//
//// SendPublicFederatedMessage will send an arbitrary provided message to followers.
//func SendPublicFederatedMessage(message string) error {
//	return outbox.SendPublicMessage(message)
//}
//
//// SendDirectFederatedMessage will send a direct message to a single account.
//func SendDirectFederatedMessage(message, account string) error {
//	return outbox.SendDirectMessageToAccount(message, account)
//}
//
//// GetFollowerCount will return the local tracked follower count.
//func GetFollowerCount() (int64, error) {
//	return persistence.GetFollowerCount()
//}
//
//// GetPendingFollowRequests will return the pending follow requests.
//func GetPendingFollowRequests() ([]models.Follower, error) {
//	return persistence.GetPendingFollowRequests()
//}
