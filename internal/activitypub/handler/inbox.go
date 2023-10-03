package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/shigde/sfu/internal/activitypub/inbox"
	"github.com/shigde/sfu/internal/activitypub/instance"
	log "github.com/sirupsen/logrus"
)

// InboxHandler handles inbound federated requests.
func GetInboxHandler(config *instance.FederationConfig, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			acceptInboxRequest(w, r, config, name)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func acceptInboxRequest(w http.ResponseWriter, r *http.Request, config *instance.FederationConfig, name string) {
	if !config.Enable {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	urlPathComponents := strings.Split(r.URL.Path, "/")
	var forLocalAccount string
	if len(urlPathComponents) == 5 {
		forLocalAccount = urlPathComponents[3]
	} else {
		log.Errorln("Unable to determine username from url path")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// The account this request is for must match the account name we have set
	// for federation.
	if forLocalAccount != name {
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
