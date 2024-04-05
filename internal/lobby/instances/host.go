package instances

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/pkg/authentication"
)

type host struct {
	actorIri   url.URL
	actorId    string
	token      string
	instanceId uuid.UUID

	liveStreamId string
	space        string
}

func newHost(actorIri url.URL, token string) *host {
	preferredName := preferredNameFromActor(actorIri)
	actorId := fmt.Sprintf("%s@%s", preferredName, actorIri.Host)
	instanceId := auth.CreateShigInstanceId(actorId)
	return &host{
		actorIri:   actorIri,
		actorId:    actorId,
		token:      token,
		instanceId: instanceId,
	}
}

func preferredNameFromActor(actorIri url.URL) string {
	path := strings.Split(actorIri.Path, "/")
	length := len(path)
	if length == 0 {
		return "shig"
	}

	return path[length-1]
}

func (h *host) GetUser() *authentication.User {
	return &authentication.User{
		UserId: h.actorId,
		Token:  h.token,
	}
}
