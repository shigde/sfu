package rtp

import "github.com/pion/webrtc/v3"

type StateEventHandler interface {
	OnConnectionStateChange(state webrtc.ICEConnectionState)
	OnNegotiationNeeded(offer webrtc.SessionDescription)
	OnChannel(dc *webrtc.DataChannel)
}
