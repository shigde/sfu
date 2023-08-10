package lobby

import (
	"context"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
)

type rtpEngineMock struct {
	conn *rtp.Endpoint
	err  error
}

func newRtpEngineMock() *rtpEngineMock {
	return &rtpEngineMock{}
}

func (e *rtpEngineMock) NewReceiverEndpoint(_ context.Context, _ webrtc.SessionDescription, _ chan<- *webrtc.TrackLocalStaticRTP) (*rtp.Endpoint, error) {
	return e.conn, e.err
}

func (e *rtpEngineMock) NewSenderEndpoint(_ context.Context, _ []*webrtc.TrackLocalStaticRTP) (*rtp.Endpoint, error) {
	return e.conn, e.err
}

func mockRtpEngineForOffer(answer *webrtc.SessionDescription) *rtpEngineMock {
	engine := newRtpEngineMock()
	engine.conn = mockConnection(answer)
	return engine
}

func mockConnection(answer *webrtc.SessionDescription) *rtp.Endpoint {
	ops := rtp.MockConnectionOps{
		Answer:         answer,
		GatherComplete: make(chan struct{}),
	}
	close(ops.GatherComplete)
	return rtp.NewMockConnection(ops)
}

var mockedAnswer = &webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "--a--"}
var mockedOffer = &webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: "--o--"}
