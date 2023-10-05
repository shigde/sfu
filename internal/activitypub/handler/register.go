package handler

import (
	"net/http"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/outbox"
	"github.com/shigde/sfu/internal/activitypub/remote"
	"golang.org/x/exp/slog"
)

func GetRegisterHandler(
	config *instance.FederationConfig,
	actorRep *models.ActorRepository,
	followRep *models.FollowRepository,
	sender *outbox.Sender,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if !config.Enable {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// @TODO put all this logic in a background worker
		instanceActor, err := actorRep.GetActorForUserName(r.Context(), config.InstanceUsername)
		if err != nil {
			slog.Error("reading local instance actor for username from db", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		instanceId := "http://localhost:9000/accounts/peertube"
		req, err := sender.GetAccountRequest(instanceActor.GetActorIri(), instanceId)
		if err != nil {
			slog.Error("building sign actor request for remote", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		remoteInstance, err := remote.FetchAccountAsActor(r.Context(), req)
		if err != nil {
			slog.Error("fetching actor from remote", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		remoteInstance, err = actorRep.Upsert(r.Context(), remoteInstance)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		follow := models.NewFollow(instanceActor, remoteInstance, config)
		follow, err = followRep.Add(r.Context(), follow)
		if err != nil {
			slog.Error("saving fallow", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := sender.SendFollowRequest(follow); err != nil {
			slog.Error("sending fallow request", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		return
	}
}
