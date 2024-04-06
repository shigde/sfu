package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type AnswerInstanceIngress struct {
	*Command
	sdp      *webrtc.SessionDescription
	option   []resources.Option
	Response *resources.WebRTC
}

func NewAnswerInstanceIngress(
	ctx context.Context,
	user uuid.UUID,
	sdp *webrtc.SessionDescription,
	option ...resources.Option,
) *AnswerInstanceIngress {
	command := NewCommand(ctx, user)
	return &AnswerInstanceIngress{
		Command:  command,
		sdp:      sdp,
		option:   option,
		Response: nil,
	}
}
func (c *AnswerInstanceIngress) Execute(session *sessions.Session) {
	answer, err := session.CreateIngressEndpoint(c.ParentCtx, c.sdp)
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
