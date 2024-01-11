package lobby

import (
	"github.com/pion/webrtc/v3"
)

type liveStreamSender interface {
	AddTrack(track webrtc.TrackLocal)
	RemoveTrack(track webrtc.TrackLocal)
}
