package activitypub

import (
	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/activitypub/crypto"
	"github.com/shigde/sfu/internal/activitypub/handler"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/outbox"
)

// StartRouter will start the federation specific http router.
func extendRouter(
	router *mux.Router,
	config *instance.FederationConfig,
	actorRep *models.ActorRepository,
	signer *crypto.Signer,
	sender *outbox.Sender,
) error {
	router.HandleFunc("/.well-known/webfinger", handler.GetWebfinger(config))

	// Single ActivityPub Actor
	router.HandleFunc("/federation/user/", handler.GetActorHandler(config, actorRep, signer))

	// Single AP object
	router.HandleFunc("/federation/", handler.GetObjectHandler(config, signer))

	// Register request for instances
	router.HandleFunc("/federation/register", handler.GetRegisterHandler(config, actorRep, sender)).Methods("GET")

	return nil
}
