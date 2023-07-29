package rtp

import (
	"github.com/pion/webrtc/v3"
)

type MockConnectionOps struct {
	Answer         *webrtc.SessionDescription
	GatherComplete chan struct{}
}

func NewMockConnection(ops MockConnectionOps) *Connection {
	conn := &Connection{}
	if ops.Answer != nil {
		conn.peerConnection = &mockPeerConnector{ops.Answer}
	}

	if ops.GatherComplete != nil {
		conn.gatherComplete = ops.GatherComplete
	}

	return conn
}

type mockPeerConnector struct {
	SDP *webrtc.SessionDescription
}

func (m *mockPeerConnector) LocalDescription() *webrtc.SessionDescription {
	return m.SDP
}

func (m *mockPeerConnector) SetRemoteDescription(_ webrtc.SessionDescription) error {
	return nil
}
func (m *mockPeerConnector) GetSenders() []*webrtc.RTPSender {
	return nil
}

func (m *mockPeerConnector) AddTrack(_ webrtc.TrackLocal) (*webrtc.RTPSender, error) {
	return nil, nil
}
