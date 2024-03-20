package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type CreateEgressResource struct {
	*command
	sdp      *webrtc.SessionDescription
	option   []resources.Option
	Response chan *resources.WebRTC
}

func NewCreateEgressResource(
	ctx context.Context,
	user uuid.UUID,
	sdp *webrtc.SessionDescription,
	option ...resources.Option,
) *CreateEgressResource {
	command := newCommand(ctx, user)
	return &CreateEgressResource{
		command:  command,
		sdp:      sdp,
		option:   option,
		Response: make(chan *resources.WebRTC),
	}
}
func (c *CreateEgressResource) Execute(ctx context.Context, session *sessions.Session) {
	answer, err := session.CreateEgressEndpoint(ctx, c.sdp)
	if err != nil {
		c.Fail(err)
		return
	}

	select {
	case <-c.ctx.Done():
	default:
		c.Response <- &resources.WebRTC{
			SDP: answer,
		}
	}
}
