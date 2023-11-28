package lobby

import "github.com/shigde/sfu/internal/rtp"

type mainStreamer interface {
	AddTrack(track *rtp.LiveTrackStaticRTP)
	RemoveTrack(id string)
	GetTracks() []*rtp.LiveTrackStaticRTP
}
