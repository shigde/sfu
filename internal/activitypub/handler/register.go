package handler

import (
	"net/http"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/outbox"
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

		//http://localhost:9000/accounts/peertube
		//http://localhost:9000/inbox
		//http://localhost:9000/accounts/peertube/inbox

		remoteInstance := &models.Actor{
			ActorIri: "http://localhost:9000/accounts/peertube",
			InboxIri: "http://localhost:9000/inbox",
		}

		if err := sender.SendFollowRequest(instanceActor, remoteInstance); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		return
	}
}
