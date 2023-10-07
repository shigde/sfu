package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/activitypub/crypto"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/request"
	"golang.org/x/exp/slog"
)

// ActorHandler handles requests for a single actor.

func GetActorHandler(
	config *instance.FederationConfig,
	actorRep *models.ActorRepository,
	signer *crypto.Signer,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if !config.Enable {
			http.Error(w, errNoFederationSupport.Error(), http.StatusMethodNotAllowed)
			return
		}

		accountName, err := getAccountName(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		actor, err := actorRep.GetActorForUserName(r.Context(), accountName)
		if err != nil {
			// User is not valid
			w.WriteHeader(http.StatusNotFound)
			return
		}

		actorIRI := actor.GetActorIri()
		publicKey := crypto.GetPublicKey(actorIRI, actor.PublicKey)
		privateKey := crypto.GetPrivateKey(actor.PrivateKey.String)

		person := models.BuildActivityApplication(actor, config)
		response := request.NewSignedResponse(signer)

		if err := response.WriteStreamResponse(person, w, publicKey, privateKey); err != nil {
			slog.Error("unable to write stream response for actor handler", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func getAccountName(r *http.Request) (string, error) {
	account, ok := mux.Vars(r)["accountName"]
	if !ok {
		return "", errAccountNameNotFound
	}
	return account, nil
}
