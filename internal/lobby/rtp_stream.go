package lobby

import (
	"github.com/pion/webrtc/v3"
)

type rtpStream struct {
	Id                     string
	audioTrack, videoTrack *webrtc.TrackLocalStaticRTP
}

func newRtpStream() *rtpStream {
	return &rtpStream{}
}
