package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type OfferInstanceEgress struct {
	*Command
	sdp      *webrtc.SessionDescription
	option   []resources.Option
	Response *resources.WebRTC
}

func NewOfferInstanceEgress(ctx context.Context, user uuid.UUID, sdp *webrtc.SessionDescription, option ...resources.Option) *OfferInstanceEgress {
	command := NewCommand(ctx, user)
	return &OfferInstanceEgress{
		Command:  command,
		sdp:      sdp,
		option:   option,
		Response: nil,
	}
}
func (c *OfferInstanceEgress) Execute(session *sessions.Session) {
	answer, err := session.CreateEgressEndpoint(c.ParentCtx, c.sdp)
	if err != nil {
		c.SetError(err)
		return
	}

	c.Response = &resources.WebRTC{
		Id:  session.Id.String(),
		SDP: answer,
	}
	c.SetDone()
}
