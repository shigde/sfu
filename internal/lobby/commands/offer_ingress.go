package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type egressRestApi interface {
	PostWhepOffer(offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error)
}

type OfferIngress struct {
	*Command
	api        egressRestApi
	signalKind sessions.SignalChannelKind
	Response   *resources.WebRTC
}

func NewOfferIngress(ctx context.Context, api egressRestApi, user uuid.UUID, signalKind sessions.SignalChannelKind) *OfferIngress {
	command := NewCommand(ctx, user)
	return &OfferIngress{
		Command:    command,
		api:        api,
		signalKind: signalKind,
		Response:   nil,
	}
}
func (c *OfferIngress) Execute(session *sessions.Session) {
	offer, err := session.OfferIngressEndpoint(c.ParentCtx, c.signalKind)
	if err != nil {
		c.SetError(err)
		return
	}

	var answer *webrtc.SessionDescription
	if answer, err = c.api.PostWhepOffer(offer); err != nil {
		c.SetError(fmt.Errorf("remote host answer request: %w", err))
		return
	}

	done := session.SetIngressAnswer(answer)
	<-done
	c.SetDone()
}
