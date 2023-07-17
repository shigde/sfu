package rtp

import "github.com/pion/webrtc/v3"

func MockConnectionWithAnswer(answer *webrtc.SessionDescription) *Connection {
	pc := &mockPeerConnector{answer}
	return &Connection{
		peerConnector: pc,
	}
}

type mockPeerConnector struct {
	SDP *webrtc.SessionDescription
}

func (m *mockPeerConnector) LocalDescription() *webrtc.SessionDescription {
	return m.SDP

}
