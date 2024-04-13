package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type OfferIngress struct {
	*Command
	signalKind sessions.SignalChannelKind
	Response   *resources.WebRTC
}

func NewOfferIngress(ctx context.Context, user uuid.UUID, signalKind sessions.SignalChannelKind) *OfferIngress {
	command := NewCommand(ctx, user)
	return &OfferIngress{
		Command:    command,
		signalKind: signalKind,
		Response:   nil,
	}
}
func (c *OfferIngress) Execute(session *sessions.Session) {
	answer, err := session.OfferIngressEndpoint(c.ParentCtx, c.signalKind)
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
