package handler

import (
	"net/http"
	"strings"

	"github.com/shigde/sfu/internal/activitypub/crypto"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/request"
	log "github.com/sirupsen/logrus"
)

// ActorHandler handles requests for a single actor.

func GetActorHandler(
	config *instance.FederationConfig,
	actorRep *models.ActorRepository,
	signer *crypto.Signer,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if !config.Enable {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		pathComponents := strings.Split(r.URL.Path, "/")
		accountName := pathComponents[3]

		actor, err := actorRep.GetActorForUserName(r.Context(), accountName)
		if err != nil {
			// User is not valid
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// If this request is for an actor's inbox then pass
		// the request to the inbox controller.
		if len(pathComponents) == 5 && pathComponents[4] == "inbox" {
			GetInboxHandler(config, accountName)(w, r)
			return
		} else if len(pathComponents) == 5 && pathComponents[4] == "outbox" {
			GetOutboxHandler(signer, actor)(w, r)
			return
			//} else if len(pathComponents) == 5 && pathComponents[4] == "followers" {
			//	// followers list
			//	FollowersHandler(w, r)
			//	return
		} else if len(pathComponents) == 5 && pathComponents[4] == "following" {
			// following list (none)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		actorIRI := actor.GetActorIri()
		publicKey := crypto.GetPublicKey(actorIRI, actor.PublicKey)
		privateKey := crypto.GetPrivateKey(actor.PrivateKey.String)

		person := models.BuildActivityPerson(actor, config)
		response := request.NewSignedResponse(signer)

		if err := response.WriteStreamResponse(person, w, publicKey, privateKey); err != nil {
			log.Errorln("unable to write stream response for actor handler", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
