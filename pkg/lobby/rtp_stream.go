package lobby

import (
	"github.com/pion/webrtc/v3"
)

type RtpStream struct {
	Id                     string `json:"Id"`
	audioTrack, videoTrack *webrtc.TrackLocalStaticRTP
	Role                   *Role
	LiveStreamId           string
	UID                    string
}

func newRtpStream(role *Role, liveStreamId string, UID string) *RtpStream {
	return &RtpStream{Role: role, LiveStreamId: liveStreamId, UID: UID}
}

func (s *RtpStream) onOffer(offer *webrtc.SessionDescription) {

}
