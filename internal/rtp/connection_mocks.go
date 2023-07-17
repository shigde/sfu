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
		conn.peerConnector = &mockPeerConnector{ops.Answer}
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
