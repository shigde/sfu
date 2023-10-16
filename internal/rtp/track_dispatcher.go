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

type TrackInfo struct {
	Kind      string
	SessionId uuid.UUID
	Track     *webrtc.TrackLocalStaticRTP
}

func newTrackInfo(id uuid.UUID, track *webrtc.TrackLocalStaticRTP, kind string) *TrackInfo {
	return &TrackInfo{
		SessionId: id,
		Track:     track,
		Kind:      kind,
	}
}

func (t *TrackInfo) GetStreamKind() string {
	return t.Kind
}

func (t *TrackInfo) GetSessionId() uuid.UUID {
	return t.SessionId
}

func (t *TrackInfo) GetTrack() *webrtc.TrackLocalStaticRTP {
	return t.Track
}
