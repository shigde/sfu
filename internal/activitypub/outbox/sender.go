package outbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func (s *Sender) SendFollowRequest(follow *models.Follow) error {
	activity, err := follow.ToAS()
	if err != nil {
		return fmt.Errorf("bilding follow activiy stream: %w", err)
	}
	b, err := models.Serialize(activity)
	if err != nil {
		return fmt.Errorf("serializing custom fediverse message activity: %w", err)
	}

	return s.SendToUser(follow.TargetActor.GetInboxIri(), b)

}

func (s *Sender) GetSignedRequest(fromActorIRI *url.URL, url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer(nil))
	if err != nil {
		return nil, fmt.Errorf("building account request object: %w", err)
	}
	ua := fmt.Sprintf("%s; https://stream.shig.de", s.config.Release)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Content-Type", "application/activity+json")

	if err := s.signer.SignRequest(req, nil, fromActorIRI); err != nil {
		return nil, fmt.Errorf("signing account request object: %w", err)
	}

	return req, nil
}

func (s *Sender) DoRequest(req *http.Request) (map[string]interface{}, error) {
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %w", err)
	}
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading boosy request: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("unmarshalling request body: %w", err)
	}

	return raw, nil

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
