package handler

import (
	"html"
	"net/http"
	"net/url"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/outbox"
	"github.com/shigde/sfu/internal/activitypub/services"
	"golang.org/x/exp/slog"
)

func GetRegisterHandler(
	config *instance.FederationConfig,
	actorService *services.ActorService,
	followRep *models.FollowRepository,
	sender *outbox.Sender,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if !config.Enable {
			http.Error(w, errNoFederationSupport.Error(), http.StatusMethodNotAllowed)
			return
		}

		accountIri, err := getAccountIriFromGet(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		instanceActor, err := actorService.GetLocalInstanceActor(r.Context())
		if err != nil {
			slog.Error("getting local instance actor", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = actorService.CreateActorFromRemoteAccount(r.Context(), accountIri.String(), instanceActor)
		if err != nil {
			slog.Error("getting remote account as actor", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//follow := models.NewFollow(instanceActor, remoteInstance, config)
		//follow, err = followRep.Add(r.Context(), follow)
		//if err != nil {
		//	slog.Error("saving fallow", "err", err)
		//	w.WriteHeader(http.StatusInternalServerError)
		//	return
		//}
		//
		//if err := sender.SendFollowRequest(follow); err != nil {
		//	slog.Error("sending fallow request", "err", err)
		//	w.WriteHeader(http.StatusInternalServerError)
		//	return
		//}
		w.WriteHeader(http.StatusCreated)
		return
	}
}

func getAccountIriFromGet(r *http.Request) (*url.URL, error) {
	iriString := r.URL.Query().Get("accountIri")
	if len(iriString) == 0 {
		return nil, errAccountIriNotFound
	}
	iri := html.UnescapeString(iriString)
	iriUrl, err := url.ParseRequestURI(iri)
	if err != nil {
		return nil, errAccountIriInvalid
	}

	return iriUrl, nil
}
