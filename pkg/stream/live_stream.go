package stream

import (
	"github.com/shigde/sfu/pkg/media"
)

type LiveStream struct {
	Id      string `json:"Id"`
	SpaceId string
	User    string
}

func (liveStream *LiveStream) EnterLobby(user string, offer *media.Offer) (*media.Answer, bool) {
	return nil, false
}
