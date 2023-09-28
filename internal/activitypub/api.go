package activitypub

import (
	"github.com/owncast/owncast/activitypub/crypto"
	"github.com/owncast/owncast/activitypub/inbox"
	"github.com/owncast/owncast/activitypub/outbox"
	"github.com/owncast/owncast/activitypub/persistence"
	"github.com/owncast/owncast/activitypub/workerpool"

	"github.com/owncast/owncast/core/data"
	"github.com/owncast/owncast/models"
	log "github.com/sirupsen/logrus"
)

type ApApi struct {
	config *FederationConfig,
	storage *storage,
}

// Start will initialize and start the federation support.
func NewApApi(config *FederationConfig, store *storage) {

	// https://stackoverflow.com/questions/69204003/insert-seed-data-at-the-first-time-of-migration-in-gorm

	persistence.Setup(datastore)
	workerpool.InitOutboundWorkerPool()
	inbox.InitInboxWorkerPool()
	StartRouter()

	// Generate the keys for signing federated activity if needed.
	if data.GetPrivateKey() == "" {
		privateKey, publicKey, err := crypto.GenerateKeys()
		_ = data.SetPrivateKey(string(privateKey))
		_ = data.SetPublicKey(string(publicKey))
		if err != nil {
			log.Errorln("Unable to get private key", err)
		}
	}
}

func (a *ApApi) boostrap()  {

}
// SendLive will send a "Go Live" message to followers.
func SendLive() error {
	return outbox.SendLive()
}

// SendPublicFederatedMessage will send an arbitrary provided message to followers.
func SendPublicFederatedMessage(message string) error {
	return outbox.SendPublicMessage(message)
}

// SendDirectFederatedMessage will send a direct message to a single account.
func SendDirectFederatedMessage(message, account string) error {
	return outbox.SendDirectMessageToAccount(message, account)
}

// GetFollowerCount will return the local tracked follower count.
func GetFollowerCount() (int64, error) {
	return persistence.GetFollowerCount()
}

// GetPendingFollowRequests will return the pending follow requests.
func GetPendingFollowRequests() ([]models.Follower, error) {
	return persistence.GetPendingFollowRequests()
}
