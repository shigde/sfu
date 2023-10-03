package activitypub

import (
	//"github.com/owncast/owncast/activitypub/controllers"
	//"github.com/owncast/owncast/router/middleware"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/activitypub/crypto"
	"github.com/shigde/sfu/internal/activitypub/handler"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
)

// StartRouter will start the federation specific http router.
func extendRouter(
	router *mux.Router,
	config *instance.FederationConfig,
	actorRep *models.ActorRepository,
	signer *crypto.Signer,
) error {
	router.HandleFunc("/.well-known/webfinger", handler.GetWebfinger(config))

	// Single ActivityPub Actor
	http.HandleFunc("/federation/user/", handler.GetActorHandler(config, actorRep, signer))

	// Single AP object
	http.HandleFunc("/federation/", handler.GetObjectHandler(config, signer))

	return nil
}
