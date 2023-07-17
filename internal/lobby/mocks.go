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

func mockRtpEngineForOffer(answer *webrtc.SessionDescription) *rtpEngineMock {
	engine := newRtpEngineMock()
	engine.conn = mockConnection(answer)
	return engine
}

func mockConnection(answer *webrtc.SessionDescription) *rtp.Connection {
	ops := rtp.MockConnectionOps{
		Answer:         answer,
		GatherComplete: make(chan struct{}),
	}
	close(ops.GatherComplete)
	return rtp.NewMockConnection(ops)
}

var mockedAnswer = &webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "--a--"}
var mockedOffer = &webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: "--o--"}
