package instances

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/auth"
	"golang.org/x/exp/slog"
)

type hostSettings struct {
	url        *url.URL
	space      string
	stream     string
	isHost     bool
	instanceId uuid.UUID
	name       string
	token      string
}

type hostx interface {
	GetHost() string
	GetSpace() string
	GetLiveStreamID() string
}

func NewHostSettings(host hostx, homeInstanceActor *url.URL, token string) *hostSettings {
	isHost := isSameShigInstance(host.GetHost(), homeInstanceActor.String())
	actorUrl, _ := url.Parse(host.GetHost())
	hostUrl, _ := url.Parse(fmt.Sprintf("%s://%s", actorUrl.Scheme, actorUrl.Host))

	// @TODO read this from the activitypub.models.Instance object
	// instanceId, _ := uuid.Parse("7251719d-a687-4f76-995c-05e03faff69d")
	// name := "shig@fosdem-stream.shig.de"
	preferredName := "shig"
	actorId := fmt.Sprintf("%s@%s", preferredName, homeInstanceActor.Host)
	instanceId := auth.CreateShigInstanceId(actorId)
	return &hostSettings{
		instanceId: instanceId,
		isHost:     isHost,
		url:        hostUrl,
		token:      token,
		name:       actorId,
		space:      host.GetSpace(),
		stream:     host.GetLiveStreamID(),
	}
}

func isSameShigInstance(streamHostActor string, instanceActorUrl string) bool {
	isSame := true
	if instanceActorUrl != streamHostActor {
		isSame = false
	}
	slog.Debug("lobby.isSameShigInstance: lobby host and home instance are same?", "isSame", isSame, "streamHostActor", streamHostActor, "instanceActor", instanceActorUrl)
	return isSame
}
