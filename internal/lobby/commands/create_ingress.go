package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type CresteIngress struct {
	*Command
	sdp        *webrtc.SessionDescription
	signalKind sessions.SignalChannelKind
	Response   *resources.WebRTC
}

func NewCreateIngress(
	ctx context.Context,
	user uuid.UUID,
	sdp *webrtc.SessionDescription,
	signalKind sessions.SignalChannelKind,
) *CresteIngress {
	command := NewCommand(ctx, user)
	return &CresteIngress{
		Command:    command,
		sdp:        sdp,
		signalKind: signalKind,
		Response:   nil,
	}
}
func (c *CresteIngress) Execute(session *sessions.Session) {
	answer, err := session.CreateIngressEndpoint(c.ParentCtx, c.sdp, c.signalKind)
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
