package rtp

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type TrackInfo struct {
	Kind      TrackInfoKind
	SessionId uuid.UUID
	Track     *webrtc.TrackLocalStaticRTP
	LiveTrack *LiveTrackStaticRTP
}

type TrackInfoKind int

const (
	TrackInfoKindGuest TrackInfoKind = iota + 1
	TrackInfoKindMain
)

func newTrackInfo(id uuid.UUID, track *webrtc.TrackLocalStaticRTP, liveTrack *LiveTrackStaticRTP, kind TrackInfoKind) *TrackInfo {
	return &TrackInfo{
		SessionId: id,
		Track:     track,
		LiveTrack: liveTrack,
		Kind:      kind,
	}
}

func (t *TrackInfo) GetStreamKind() TrackInfoKind {
	return t.Kind
}

func (t *TrackInfo) GetSessionId() uuid.UUID {
	return t.SessionId
}

func (t *TrackInfo) GetTrack() *webrtc.TrackLocalStaticRTP {
	return t.Track
}

func (t *TrackInfo) GetLiveTrack() *LiveTrackStaticRTP {
	return t.LiveTrack
}
