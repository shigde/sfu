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

func (e *rtpEngineMock) EstablishIngressEndpoint(_ context.Context, _ uuid.UUID, _ webrtc.SessionDescription, _ rtp.TrackDispatcher, _ rtp.StateEventHandler) (*rtp.Endpoint, error) {
	return e.conn, e.err
}

func (e *rtpEngineMock) EstablishEgressEndpoint(_ context.Context, _ uuid.UUID, _ []webrtc.TrackLocal, _ rtp.StateEventHandler) (*rtp.Endpoint, error) {
	return e.conn, e.err
}

func (e *rtpEngineMock) EstablishStaticEgressEndpoint(_ context.Context, _ uuid.UUID, _ webrtc.SessionDescription, _ ...rtp.EndpointOption) (*rtp.Endpoint, error) {
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

type mainStreamerMock struct {
	Tracks map[string]*rtp.LiveTrackStaticRTP
}

func newMainStreamerMock() *mainStreamerMock {
	tracks := make(map[string]*rtp.LiveTrackStaticRTP)
	return &mainStreamerMock{
		Tracks: tracks,
	}
}
func (sf *mainStreamerMock) AddTrack(track *rtp.LiveTrackStaticRTP) {
	sf.Tracks[track.ID()] = track
}

func (sf *mainStreamerMock) RemoveTrack(track *rtp.LiveTrackStaticRTP) {
	delete(sf.Tracks, track.ID())
}
func (sf *mainStreamerMock) GetTracks() []*rtp.LiveTrackStaticRTP {
	tracks := make([]*rtp.LiveTrackStaticRTP, len(sf.Tracks))
	for _, track := range sf.Tracks {
		tracks = append(tracks, track)
	}
	return tracks
}
