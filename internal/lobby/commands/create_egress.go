package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type CreateEgress struct {
	*Command
	sdp        *webrtc.SessionDescription
	signalKind sessions.SignalChannelKind
	Response   *resources.WebRTC
}

func NewCreateEgress(ctx context.Context, user uuid.UUID, sdp *webrtc.SessionDescription, signalKind sessions.SignalChannelKind) *CreateEgress {
	command := NewCommand(ctx, user)
	return &CreateEgress{
		Command:    command,
		sdp:        sdp,
		signalKind: signalKind,
		Response:   nil,
	}
}
func (c *CreateEgress) Execute(session *sessions.Session) {
	answer, err := session.CreateEgressEndpoint(c.ParentCtx, c.sdp, c.signalKind)
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
