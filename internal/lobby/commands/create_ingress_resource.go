package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type CreateIngressResource struct {
	*command
	sdp      *webrtc.SessionDescription
	option   []resources.Option
	Response chan *resources.WebRTC
}

func NewCreateIngressResource(
	ctx context.Context,
	user uuid.UUID,
	sdp *webrtc.SessionDescription,
	option ...resources.Option,
) *CreateIngressResource {
	command := newCommand(ctx, user)
	return &CreateIngressResource{
		command:  command,
		sdp:      sdp,
		option:   option,
		Response: make(chan *resources.WebRTC),
	}
}
func (c *CreateIngressResource) Execute(ctx context.Context, session *sessions.Session) {
	answer, err := session.CreateIngressEndpoint(ctx, c.sdp)
	if err != nil {
		c.Fail(err)
		return
	}

	select {
	case <-c.ctx.Done():
	default:
		c.Response <- &resources.WebRTC{
			Id:  session.Id.String(),
			SDP: answer,
		}
	}
}
