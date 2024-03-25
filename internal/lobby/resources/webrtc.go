package resources

import "github.com/pion/webrtc/v3"

type WebRTC struct {
	Id  string
	SDP *webrtc.SessionDescription
}
