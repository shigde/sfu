package outbox

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/shigde/sfu/internal/activitypub/crypto"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/remote"
	"github.com/shigde/sfu/internal/activitypub/webfinger"
	"github.com/shigde/sfu/internal/activitypub/workerpool"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/teris-io/shortid"
	"golang.org/x/exp/slog"
)

type Sender struct {
	config          *instance.FederationConfig
	webfingerClient *webfinger.Client
	resolver        *remote.Resolver
	signer          *crypto.Signer
}

func NewSender(
	config *instance.FederationConfig,
	webfingerClient *webfinger.Client,
	resolver *remote.Resolver,
	signer *crypto.Signer,
) *Sender {
	return &Sender{
		config,
		webfingerClient,
		resolver,
		signer,
	}
}

func (s *Sender) SendFollowRequest(actor *models.Actor, target *models.Actor) error {
	follow := models.NewFollow(actor, target, s.config)
	activity, _ := follow.ToAS(context.Background())
	b, err := models.Serialize(activity)
	if err != nil {
		return fmt.Errorf("serializing custom fediverse message activity: %w", err)
	}

	return s.SendToUser(target.GetInboxIri(), b)

}

func (s *Sender) GetAccountRequest(fromActorIRI *url.URL, url string) (*http.Request, error) {
	req, _ := http.NewRequest(http.MethodGet, url, bytes.NewBuffer(nil))
	ua := fmt.Sprintf("%s; https://stream.shig.de", s.config.Release)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Content-Type", "application/activity+json")

	if err := s.signer.SignRequest(req, nil, fromActorIRI); err != nil {
		slog.Error("error signing request:", "err", err)
		return nil, err
	}

	return req, nil
}

func (s *Sender) createBaseOutboundMessage(textContent string) (vocab.ActivityStreamsCreate, string, vocab.ActivityStreamsNote, string) {
	localActor := instance.BuildAccountIri(s.config.InstanceUrl, s.config.InstanceUsername)
	noteID := shortid.MustGenerate()
	noteIRI := instance.BuildResourceIri(s.config.InstanceUrl, noteID)
	id := shortid.MustGenerate()
	activity := models.CreateCreateActivity(id, s.config.InstanceUrl, localActor)
	object := streams.NewActivityStreamsObjectProperty()
	activity.SetActivityStreamsObject(object)

	note := models.MakeNote(textContent, noteIRI, localActor)
	object.AppendActivityStreamsNote(note)

	return activity, id, note, noteID
}

func (s *Sender) SendToUser(inbox *url.URL, payload []byte) error {
	localActor := instance.BuildAccountIri(s.config.InstanceUrl, s.config.InstanceUsername)

	req, err := s.createSignedRequest(payload, inbox, localActor)
	if err != nil {
		return errors.Wrap(err, "unable to create outbox request")
	}

	workerpool.AddToOutboundQueue(req)

	return nil
}

func (s *Sender) createSignedRequest(payload []byte, url *url.URL, fromActorIRI *url.URL) (*http.Request, error) {
	slog.Debug("Sending", "payload", string(payload), "toUrl", url)

	req, _ := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(payload))

	ua := fmt.Sprintf("%s; https://stream.shig.de", s.config.Release)

	req.Header.Set("User-Agent", ua)
	req.Header.Set("Content-Type", "application/activity+json")

	if err := s.signer.SignRequest(req, payload, fromActorIRI); err != nil {
		slog.Error("error signing request:", "err", err)
		return nil, err
	}

	return req, nil
}
