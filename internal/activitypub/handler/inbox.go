package handler

import (
	"io"
	"net/http"

	"github.com/shigde/sfu/internal/activitypub/inbox"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slog"
)

// InboxHandler handles inbound federated requests.
func GetInboxHandler(
	config *instance.FederationConfig,
	actorRep *models.ActorRepository,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			acceptInboxRequest(w, r, config, actorRep)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func acceptInboxRequest(w http.ResponseWriter, r *http.Request, config *instance.FederationConfig, actorRep *models.ActorRepository) {
	if !config.Enable {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	forLocalAccount, err := getAccountName(r)
	if err != nil {
		slog.Error("unable to determine username from url path in inbox handler")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// The account this request is for must match the account name we have set
	// for federation.
	_, err = actorRep.GetActorForUserName(r.Context(), forLocalAccount)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorln("Unable to read inbox request payload", err)
		return
	}

	inboxRequest := inbox.InboxRequest{Request: r, ForLocalAccount: forLocalAccount, Body: data}
	inbox.AddToQueue(inboxRequest)
	w.WriteHeader(http.StatusAccepted)
}
