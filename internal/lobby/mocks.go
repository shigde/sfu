package lobby

import (
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
)

type rtpEngineMock struct {
	conn *rtp.Connection
	err  error
}

func newRtpEngineMock() *rtpEngineMock {
	return &rtpEngineMock{}
}

func (e *rtpEngineMock) NewConnection(_ webrtc.SessionDescription, _ string) (*rtp.Connection, error) {
	return e.conn, e.err
}

func mockConnection(answer *webrtc.SessionDescription) *rtp.Connection {
	ops := rtp.MockConnectionOps{
		Answer:         answer,
		GatherComplete: make(chan struct{}),
	}
	close(ops.GatherComplete)
	return rtp.NewMockConnection(ops)
}
