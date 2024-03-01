package activitypub

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/activitypub/crypto"
	"github.com/shigde/sfu/internal/activitypub/inbox"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/outbox"
	"github.com/shigde/sfu/internal/activitypub/remote"
	"github.com/shigde/sfu/internal/activitypub/services"
	"github.com/shigde/sfu/internal/activitypub/webfinger"
	"github.com/shigde/sfu/internal/activitypub/workerpool"
	"github.com/shigde/sfu/internal/migration"
	"github.com/shigde/sfu/internal/storage"
	"github.com/superseriousbusiness/activity/pub"
)

type ApApi struct {
	config       *instance.FederationConfig
	Storage      storage.Storage
	actorRepo    *models.ActorRepository
	followRepo   *models.FollowRepository
	videoRepo    *models.VideoRepository
	actor        pub.FederatingActor
	signer       *crypto.Signer
	sender       *outbox.Sender
	actorService *services.ActorService
	videoService *services.VideoService
}

func NewApApi(
	config *instance.FederationConfig,
	storage storage.Storage,
	streamService services.StreamService,
) (*ApApi, error) {
	if err := migration.Migrate(config, storage); err != nil {
		return nil, fmt.Errorf("creation schema for federation: %w", err)
	}

	actorRepo := models.NewActorRepository(config, storage)
	followRepo := models.NewFollowRepository(config, storage)
	videoRepo := models.NewVideoRepository(config, storage)
	instanceRepo := models.NewInstanceRepository(config, storage)

	// @TODO this is a skeleton, please use this as blueprint to clean up the source
	// @TODO currently we follow the implementation from Owncast, which is little tricky but was faster to implement
	behavior := NewCommonBehavior()
	protocol := NewFederatingProtocol()
	database := NewDatabase(actorRepo, followRepo)
	clock := NewClock()

	actor := pub.NewFederatingActor(behavior, protocol, database, clock)

	signer := crypto.NewSigner(actorRepo)

	webfingerClient := webfinger.NewClient(config)

	resolver := remote.NewResolver(config, signer)

	sender := outbox.NewSender(config, webfingerClient, resolver, signer)

	actorService := services.NewActorService(config, actorRepo, sender)

	instService := services.NewInstanceService(config, instanceRepo)

	videoService := services.NewVideoService(config, actorService, videoRepo, streamService, instService)

	return &ApApi{
		config:       config,
		Storage:      storage,
		actorRepo:    actorRepo,
		followRepo:   followRepo,
		videoRepo:    videoRepo,
		actor:        actor,
		signer:       signer,
		sender:       sender,
		actorService: actorService,
		videoService: videoService,
	}, nil

}

func (a *ApApi) BoostrapApi(router *mux.Router) error {
	if err := extendRouter(router, a.config, a.actorRepo, a.followRepo, a.signer, a.sender, a.actorService); err != nil {
		return fmt.Errorf("extending router with federation endpoints: %w", err)
	}

	if a.config.Enable {
		resolver := remote.NewResolver(a.config, a.signer)
		workerpool.InitOutboundWorkerPool()
		inbox.InitInboxWorkerPool(a.followRepo, a.videoService, resolver)
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
