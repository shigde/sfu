package lobby

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
)

var (
	mockedAnswer                = &webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "--a--"}
	mockedOffer                 = &webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: "--o--"}
	onQuitSessionInternallyStub = func(ctx context.Context, user uuid.UUID) bool {
		return true
	}
)

type rtpEngineMock struct {
	conn *rtp.Endpoint
	err  error
}

func newRtpEngineMock() *rtpEngineMock {
	return &rtpEngineMock{}
}

func (e *rtpEngineMock) NewReceiverEndpoint(_ context.Context, _ webrtc.SessionDescription, _ rtp.TrackDispatcher, _ rtp.StateEventHandler) (*rtp.Endpoint, error) {
	return e.conn, e.err
}

func (e *rtpEngineMock) NewSenderEndpoint(_ context.Context, _ []*webrtc.TrackLocalStaticRTP, _ rtp.StateEventHandler) (*rtp.Endpoint, error) {
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

func mockIdelConnection() *rtp.Endpoint {
	ops := rtp.MockConnectionOps{
		Answer:         nil,
		GatherComplete: make(chan struct{}),
	}
	return rtp.NewMockConnection(ops)
}
