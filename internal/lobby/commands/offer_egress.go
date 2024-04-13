package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type ingressRestApi interface {
	PostWhipOffer(offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error)
}

type OfferEgress struct {
	*Command
	api        ingressRestApi
	signalKind sessions.SignalChannelKind
	Response   *resources.WebRTC
}

func NewOfferEgress(ctx context.Context, api ingressRestApi, user uuid.UUID, signalKind sessions.SignalChannelKind) *OfferEgress {
	command := NewCommand(ctx, user)
	return &OfferEgress{
		Command:    command,
		api:        api,
		signalKind: signalKind,
		Response:   nil,
	}
}
func (c *OfferEgress) Execute(session *sessions.Session) {
	offer, err := session.OfferEgressEndpoint(c.ParentCtx, c.signalKind)
	if err != nil {
		c.SetError(err)
		return
	}

	var answer *webrtc.SessionDescription
	if answer, err = c.api.PostWhipOffer(offer); err != nil {
		c.SetError(fmt.Errorf("remote host answer request: %w", err))
		return
	}

	done := session.SetEgressAnswer(answer)
	<-done
	c.SetDone()
}
