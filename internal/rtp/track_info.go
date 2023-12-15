package rtp

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type TrackInfo struct {
	Purpose   Purpose
	SessionId uuid.UUID
	Track     *webrtc.TrackLocalStaticRTP
	LiveTrack *LiveTrackStaticRTP
}

type Purpose int

const (
	PurposeGuest Purpose = iota + 1
	PurposeMain
)

func newTrackInfo(id uuid.UUID, track *webrtc.TrackLocalStaticRTP, liveTrack *LiveTrackStaticRTP, purpose Purpose) *TrackInfo {
	return &TrackInfo{
		SessionId: id,
		Track:     track,
		LiveTrack: liveTrack,
		Purpose:   purpose,
	}
}

func (t *TrackInfo) GetPurpose() Purpose {
	return t.Purpose
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

func (t *TrackInfo) GetTrackLocal() webrtc.TrackLocal {
	if t.Purpose == PurposeMain {
		return t.LiveTrack
	}
	return t.Track
}
