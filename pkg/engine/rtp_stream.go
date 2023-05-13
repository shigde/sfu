package engine

import "github.com/pion/webrtc/v3"

type RtpStream struct {
	Id                     string `json:"Id"`
	audioTrack, videoTrack *webrtc.TrackLocalStaticRTP
}
