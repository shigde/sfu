package rtp

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type TrackInfo struct {
	Purpose   Purpose
	SessionId uuid.UUID
	Track     *webrtc.TrackLocalStaticRTP
}

type Purpose int

const (
	PurposeGuest Purpose = iota + 1
	PurposeMain
)

func (p Purpose) ToString() string {
	switch p {
	case PurposeGuest:
		return "guest"
	case PurposeMain:
		return "main"
	default:
		return "guest"
	}
}

func newTrackInfo(id uuid.UUID, track *webrtc.TrackLocalStaticRTP, purpose Purpose) *TrackInfo {
	return &TrackInfo{
		SessionId: id,
		Track:     track,
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

func (t *TrackInfo) GetTrackLocal() webrtc.TrackLocal {
	return t.Track
}
