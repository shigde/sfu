package rtp

import "github.com/pion/webrtc/v3"

type mediaStateEventHandler struct {
}

func newMediaStateEventHandler() *mediaStateEventHandler {
	return &mediaStateEventHandler{}
}

func (h *mediaStateEventHandler) OnConnectionStateChange(state webrtc.ICEConnectionState) {
}
func (h *mediaStateEventHandler) OnNegotiationNeeded(offer webrtc.SessionDescription) {}
func (h *mediaStateEventHandler) OnChannel(dc *webrtc.DataChannel)                    {}
