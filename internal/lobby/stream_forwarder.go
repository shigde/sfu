package lobby

import "github.com/shigde/sfu/internal/rtp"

type streamForwarder interface {
	AddTrack(track *rtp.LiveTrackStaticRTP)
}
