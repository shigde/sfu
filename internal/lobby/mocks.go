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

func (e *rtpEngineMock) EstablishIngressEndpoint(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ webrtc.SessionDescription, _ ...rtp.EndpointOption) (*rtp.Endpoint, error) {
	return e.conn, e.err
}

func (e *rtpEngineMock) EstablishEgressEndpoint(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ ...rtp.EndpointOption) (*rtp.Endpoint, error) {
	return e.conn, e.err
}

func (e *rtpEngineMock) EstablishStaticEgressEndpoint(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ webrtc.SessionDescription, _ ...rtp.EndpointOption) (*rtp.Endpoint, error) {
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

type liveStreamSenderMock struct {
	Tracks map[string]webrtc.TrackLocal
}

func newLiveStreamSenderMock() *liveStreamSenderMock {
	tracks := make(map[string]webrtc.TrackLocal)
	return &liveStreamSenderMock{
		Tracks: tracks,
	}
}
func (sf *liveStreamSenderMock) AddTrack(track webrtc.TrackLocal) {
	sf.Tracks[track.ID()] = track
}

func (sf *liveStreamSenderMock) RemoveTrack(track webrtc.TrackLocal) {
	delete(sf.Tracks, track.ID())
}
func (sf *liveStreamSenderMock) GetTracks() []webrtc.TrackLocal {
	tracks := make([]webrtc.TrackLocal, len(sf.Tracks))
	for _, track := range sf.Tracks {
		tracks = append(tracks, track)
	}
	return tracks
}
