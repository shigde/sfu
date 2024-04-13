package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type OfferEgress struct {
	*Command
	signalKind sessions.SignalChannelKind
	Response   *resources.WebRTC
}

func NewOfferEgress(ctx context.Context, user uuid.UUID, signalKind sessions.SignalChannelKind) *OfferEgress {
	command := NewCommand(ctx, user)
	return &OfferEgress{
		Command:    command,
		signalKind: signalKind,
		Response:   nil,
	}
}
func (c *OfferEgress) Execute(session *sessions.Session) {
	answer, err := session.OfferEgressEndpoint(c.ParentCtx, c.signalKind)
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
