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
	router.HandleFunc("/.well-known/webfinger", handler.GetWebfinger(config)).Methods("GET")

	// Single ActivityPub Actor
	router.HandleFunc("/federation/accounts/{accountName}", handler.GetActorHandler(config, actorRep, signer)).Methods("GET")
	router.HandleFunc("/federation/accounts/{accountName}/inbox", handler.GetInboxHandler(config, actorRep)).Methods("POST")

	//outbox, followers, following

	// Single AP object
	router.HandleFunc("/federation/", handler.GetObjectHandler(config, signer))

	// Register request for instances
	router.HandleFunc("/federation/register", handler.GetRegisterHandler(config, actorRep, sender)).Methods("GET")

	return nil
}
