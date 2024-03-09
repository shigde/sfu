package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type CreateIngressResourceCommand struct {
	ctx      context.Context
	user     uuid.UUID
	sdp      *webrtc.SessionDescription
	option   []resources.Option
	Response chan *resources.WebRTC
	Err      chan error
}

func NewCreateIngressResourceCommand(
	ctx context.Context,
	user uuid.UUID,
	sdp *webrtc.SessionDescription,
	option ...resources.Option,
) *CreateIngressResourceCommand {
	return &CreateIngressResourceCommand{
		ctx:      ctx,
		user:     user,
		sdp:      sdp,
		option:   option,
		Response: make(chan *resources.WebRTC),
		Err:      make(chan error),
	}
}
func (c *CreateIngressResourceCommand) Execute(session *sessions.Session) {
	// go session.do(c)
}
