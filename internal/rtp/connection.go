package rtp

import "github.com/pion/webrtc/v3"

type Connection struct {
	pc       *webrtc.PeerConnection
	receiver *receiver
	sender   *sender
	answer   chan *webrtc.SessionDescription
	closed   chan struct{}
}
