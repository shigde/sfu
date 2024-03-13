package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type CreateIngressResource struct {
	ctx      context.Context
	user     uuid.UUID
	sdp      *webrtc.SessionDescription
	option   []resources.Option
	Response chan *resources.WebRTC
	Err      chan error
}

func (c *CreateIngressResource) GetUserId() uuid.UUID {
	return c.user
}

func (c *CreateIngressResource) Fail(err error) {
	select {
	case <-c.ctx.Done():
	default:
		c.Err <- err
	}
}

func NewCreateIngressResource(
	ctx context.Context,
	user uuid.UUID,
	sdp *webrtc.SessionDescription,
	option ...resources.Option,
) *CreateIngressResource {
	return &CreateIngressResource{
		ctx:      ctx,
		user:     user,
		sdp:      sdp,
		option:   option,
		Response: make(chan *resources.WebRTC),
		Err:      make(chan error),
	}
}
func (c *CreateIngressResource) Execute(ctx context.Context, session *sessions.Session) {
	answer, err := session.CreateIngressEndpoint(ctx, c.sdp)
	if err != nil {
		c.Fail(err)
		return
	}

	c.Response <- &resources.WebRTC{
		SDP: answer,
	}
}
