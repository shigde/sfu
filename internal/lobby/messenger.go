package lobby

import "github.com/pion/webrtc/v3"

type messenger struct {
	dc *webrtc.DataChannel
}

func newMessenger(dc *webrtc.DataChannel) *messenger {
	return &messenger{
		dc: dc,
	}
}

func (m *messenger) sendOffer(_ webrtc.SessionDescription) error {
	return nil
}
