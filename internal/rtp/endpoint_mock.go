package rtp

import (
	"context"

	"github.com/pion/webrtc/v3"
)

type MockConnectionOps struct {
	Answer         *webrtc.SessionDescription
	GatherComplete chan struct{}
}

func NewMockConnection(ops MockConnectionOps) *Endpoint {
	conn := &Endpoint{
		sessionCxt: context.Background(),
	}
	if ops.Answer != nil {
		conn.peerConnection = &mockPeerConnector{SDP: ops.Answer}
	}

	if ops.GatherComplete != nil {
		conn.gatherComplete = ops.GatherComplete
	}
	conn.initComplete = make(chan struct{})
	return conn
}

type mockPeerConnector struct {
	SDP       *webrtc.SessionDescription
	RTPSender []*webrtc.RTPSender
}

func (m *mockPeerConnector) LocalDescription() *webrtc.SessionDescription {
	return m.SDP
}
func (m *mockPeerConnector) SetLocalDescription(_ webrtc.SessionDescription) error { return nil }
func (m *mockPeerConnector) SetRemoteDescription(_ webrtc.SessionDescription) error {
	return nil
}
func (m *mockPeerConnector) GetSenders() []*webrtc.RTPSender {
	return m.RTPSender
}
func (m *mockPeerConnector) GetTransceivers() []*webrtc.RTPTransceiver { return nil }
func (m *mockPeerConnector) AddTrack(_ webrtc.TrackLocal) (*webrtc.RTPSender, error) {
	return nil, nil
}
func (m *mockPeerConnector) RemoveTrack(_ *webrtc.RTPSender) error {
	return nil
}
func (m *mockPeerConnector) SignalingState() webrtc.SignalingState {
	return webrtc.SignalingStateStable
}
func (m *mockPeerConnector) OnICEConnectionStateChange(f func(webrtc.ICEConnectionState)) {}
func (m *mockPeerConnector) OnNegotiationNeeded(f func())                                 {}
func (m *mockPeerConnector) CreateOffer(_ *webrtc.OfferOptions) (webrtc.SessionDescription, error) {
	return webrtc.SessionDescription{}, nil
}
func (m *mockPeerConnector) CreateAnswer(options *webrtc.AnswerOptions) (webrtc.SessionDescription, error) {
	return webrtc.SessionDescription{}, nil
}

func (m *mockPeerConnector) Close() error {
	return nil
}
