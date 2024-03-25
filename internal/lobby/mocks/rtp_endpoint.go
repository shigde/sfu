package mocks

import (
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
)

func NewEndpoint(answer *webrtc.SessionDescription) *rtp.Endpoint {
	ops := rtp.MockConnectionOps{
		Answer:         answer,
		GatherComplete: make(chan struct{}),
	}
	close(ops.GatherComplete)
	return rtp.NewMockConnection(ops)
}

func NewIdelEndpoint() *rtp.Endpoint {
	ops := rtp.MockConnectionOps{
		Answer:         nil,
		GatherComplete: make(chan struct{}),
	}
	return rtp.NewMockConnection(ops)
}
