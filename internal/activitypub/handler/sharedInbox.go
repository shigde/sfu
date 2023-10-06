package handler

import (
	"io"
	"net/http"

	"github.com/shigde/sfu/internal/activitypub/inbox"
	"github.com/shigde/sfu/internal/activitypub/instance"
	log "github.com/sirupsen/logrus"
)

// InboxHandler handles inbound federated requests.
func GetSharedInboxHandler(
	config *instance.FederationConfig,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			acceptSharedInboxRequest(w, r, config)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func acceptSharedInboxRequest(w http.ResponseWriter, r *http.Request, config *instance.FederationConfig) {
	if !config.Enable {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorln("Unable to read inbox request payload", err)
		return
	}

	inboxRequest := inbox.InboxRequest{Request: r, ForLocalAccount: config.InstanceUsername, Body: data}
	inbox.AddToQueue(inboxRequest)
	w.WriteHeader(http.StatusAccepted)
}
