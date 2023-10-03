package handler

import (
	"net/http"

	"github.com/shigde/sfu/internal/activitypub/crypto"
	"github.com/shigde/sfu/internal/activitypub/instance"
)

// ObjectHandler handles requests for a single federated ActivityPub object.
func GetObjectHandler(
	config *instance.FederationConfig,
	signer *crypto.Signer,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !config.Enable {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// If private federation mode is enabled do not allow access to objects.
		if config.IsPrivate {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		//iri := strings.Join([]string{strings.TrimSuffix(config.InstanceUrl.String(), "/"), r.URL.Path}, "")

		//object, _, _, err := persistence.GetObjectByIRI(iri)
		//if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
		//}
		//
		//accountName := data.GetDefaultFederationUsername()
		//actorIRI := apmodels.MakeLocalIRIForAccount(accountName)
		//publicKey := crypto.GetPublicKey(actorIRI)
		//
		//if err := requests.WriteResponse([]byte(object), w, publicKey); err != nil {
		//	slog.Error(err)
		//}
	}
}
