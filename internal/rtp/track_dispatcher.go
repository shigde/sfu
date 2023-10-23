package rtp

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type TrackDispatcher interface {
	DispatchAddTrack(track *TrackInfo)
	DispatchRemoveTrack(track *TrackInfo)
}

type DispatchTrack interface {
	GetStreamKind() string
	GetSessionId() uuid.UUID
	GetTrack() *webrtc.TrackLocalStaticRTP
}
