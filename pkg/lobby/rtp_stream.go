package lobby

import (
	"github.com/pion/webrtc/v3"
)

type rtpStream struct {
	Id                     string `json:"Id"`
	audioTrack, videoTrack *webrtc.TrackLocalStaticRTP
	Role                   *Role
	LiveStreamId           string
	UID                    string
}

func newRtpStream(role *Role, liveStreamId string, UID string) *rtpStream {
	return &rtpStream{Role: role, LiveStreamId: liveStreamId, UID: UID}
}
