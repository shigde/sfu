package rtp

import "github.com/pion/webrtc/v3"

type TrackDispatcher interface {
	DispatchAddTrack(track *webrtc.TrackLocalStaticRTP)
	DispatchRemoveTrack(track *webrtc.TrackLocalStaticRTP)
}
