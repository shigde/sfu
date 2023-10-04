package handler

import (
	"net/http"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/outbox"
	"github.com/shigde/sfu/internal/activitypub/remote"
)

func GetRegisterHandler(
	config *instance.FederationConfig,
	actorRep *models.ActorRepository,
	sender *outbox.Sender,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if !config.Enable {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		instanceActor, err := actorRep.GetActorForUserName(r.Context(), config.InstanceUsername)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		instanceId := "http://localhost:9000/accounts/peertube"
		req, _ := sender.GetAccountRequest(instanceActor.GetActorIri(), instanceId)

		remoteInstance, err := remote.FetchAccountAsActor(r.Context(), req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		remoteInstance, err = actorRep.Upsert(r.Context(), remoteInstance)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := sender.SendFollowRequest(instanceActor, remoteInstance); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		return
	}
}
