package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/prometheus/common/log"
	"github.com/shigde/sfu/internal/activitypub"
	"github.com/shigde/sfu/internal/activitypub/apmodels"
)

func GetWebfinger(config *activitypub.FederationConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !config.Enable {
			w.WriteHeader(http.StatusMethodNotAllowed)
			log.Debugln("webfinger request rejected! Federation is not enabled")
			return
		}

		resource := r.URL.Query().Get("resource")
		preAcct, account, foundAcct := strings.Cut(resource, "acct:")

		if !foundAcct || preAcct != "" {
			w.WriteHeader(http.StatusBadRequest)
			log.Debugln("webfinger request rejected! Malformed resource in query: " + resource)
			return
		}

		userComponents := strings.Split(account, "@")
		if len(userComponents) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			log.Debugln("webfinger request rejected! Malformed account in query: " + account)
			return
		}
		host := userComponents[1]
		user := userComponents[0]

		if _, valid := data.GetFederatedInboxMap()[user]; !valid {
			w.WriteHeader(http.StatusNotFound)
			log.Debugln("webfinger request rejected! Invalid user: " + user)
			return
		}

		// If the webfinger request doesn't match our server then it
		// should be rejected.
		if config.InstanceHostname != host {
			w.WriteHeader(http.StatusNotImplemented)
			log.Debugln("webfinger request rejected! Invalid query host: " + host + " instanceHostString: " + config.InstanceHostname)
			return
		}

		webfingerResponse := apmodels.MakeWebfingerResponse(user, user, host)

		w.Header().Set("Content-Type", "application/jrd+json")

		if err := json.NewEncoder(w).Encode(webfingerResponse); err != nil {
			log.Errorln("unable to write webfinger response", err)
		}
	}
}
