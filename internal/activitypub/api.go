package activitypub

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
)

type ApApi struct {
	Config           *instance.FederationConfig
	InstanceProperty *instance.Property
	Storage          instance.Storage
	actorRepo        *models.ActorRepository
}

func NewApApi(config *instance.FederationConfig, storage instance.Storage) (*ApApi, error) {
	instanceProperty := instance.NewProperty(config)

	actorRepo, err := models.NewActorRepository(instanceProperty, storage)
	if err != nil {
		return nil, fmt.Errorf("creation actor repository: %w", err)
	}

	return &ApApi{
		Config:           config,
		InstanceProperty: instanceProperty,
		Storage:          storage,
		actorRepo:        actorRepo,
	}, nil

	// https://stackoverflow.com/questions/69204003/insert-seed-data-at-the-first-time-of-migration-in-gorm

	//persistence.Setup(datastore)
	//workerpool.InitOutboundWorkerPool()
	//inbox.InitInboxWorkerPool()
	//StartRouter()

	// Generate the keys for signing federated activity if needed.

}

func (a *ApApi) BoostrapApi(router *mux.Router) error {
	if err := extendRouter(router, a.Config); err != nil {
		return fmt.Errorf("extending router with federation endpoints: %w", err)
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
