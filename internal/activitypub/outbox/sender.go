package outbox

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
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
	property        *instance.Property
	remoteWebfinger *webfinger.Client
	resolver        *remote.Resolver
	signer          *crypto.Signer
}

func NewSender(
	config *instance.FederationConfig,
	property *instance.Property,
	remoteWebfinger *webfinger.Client,
	resolver *remote.Resolver,
	signer *crypto.Signer,
) *Sender {
	return &Sender{
		config,
		property,
		remoteWebfinger,
		resolver,
		signer,
	}
}

func (s *Sender) SendDirectMessageToAccount(textContent, account string) error {
	links, err := s.remoteWebfinger.GetWebfingerLinks(account)
	if err != nil {
		return fmt.Errorf("geting webfinger links when sending private message: %w", err)
	}
	user := s.remoteWebfinger.MakeWebFingerRequestResponseFromData(links)

	iri := user.Self
	actor, err := s.resolver.GetResolvedActorFromIRI(iri)
	if err != nil {
		return fmt.Errorf("resolving actor to send message to: %w", err)
	}

	activity, _, note, _ := s.createBaseOutboundMessage(textContent)

	actorIri := actor.GetActorIri()

	// Set direct message visibility
	activity = models.MakeActivityDirect(activity, actorIri)
	note = models.MakeNoteDirect(note, actorIri)
	object := activity.GetActivityStreamsObject()
	object.SetActivityStreamsNote(0, note)

	b, err := models.Serialize(activity)
	if err != nil {
		log.Errorln("unable to serialize custom fediverse message activity", err)
		return errors.Wrap(err, "unable to serialize custom fediverse message activity")
	}

	return s.SendToUser(actor.GetInboxIri(), b)
}

func (s *Sender) createBaseOutboundMessage(textContent string) (vocab.ActivityStreamsCreate, string, vocab.ActivityStreamsNote, string) {
	localActor := instance.BuildAccountIri(s.property.InstanceUrl, s.property.InstanceUsername)
	noteID := shortid.MustGenerate()
	noteIRI := instance.BuildResourceIri(s.property.InstanceUrl, noteID)
	id := shortid.MustGenerate()
	activity := models.CreateCreateActivity(id, s.property.InstanceUrl, localActor)
	object := streams.NewActivityStreamsObjectProperty()
	activity.SetActivityStreamsObject(object)

	note := models.MakeNote(textContent, noteIRI, localActor)
	object.AppendActivityStreamsNote(note)

	return activity, id, note, noteID
}

func (s *Sender) SendToUser(inbox *url.URL, payload []byte) error {
	localActor := instance.BuildAccountIri(s.property.InstanceUrl, s.property.InstanceUsername)

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
