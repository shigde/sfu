package rtp

import (
	"github.com/google/uuid"
)

type TrackSdpInfo struct {
	Id uuid.UUID
	// source ----------
	SessionId      uuid.UUID
	IngressMid     string
	IngressTrackId string

	// sink ------------
	EgressMid     string
	EgressTrackId string

	Purpose Purpose
	Mute    bool
	Info    string
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

func newTrackSdpInfo(sessionId uuid.UUID) *TrackSdpInfo {
	id := uuid.New()
	return &TrackSdpInfo{Id: id, SessionId: sessionId, Purpose: PurposeGuest, Mute: false, Info: "Guest"}
}
