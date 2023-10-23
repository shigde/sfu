package media

import (
	"github.com/pion/webrtc/v3"
)

type EmptyStateHandler struct {
}

func (h *EmptyStateHandler) OnConnectionStateChange(state webrtc.ICEConnectionState) {
}
func (h *EmptyStateHandler) OnNegotiationNeeded(offer webrtc.SessionDescription) {}
func (h *EmptyStateHandler) OnChannel(dc *webrtc.DataChannel)                    {}
