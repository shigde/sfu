package rtp

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type TrackInfo struct {
	Kind        TrackInfoKind
	SessionId   uuid.UUID
	Track       *webrtc.TrackLocalStaticRTP
	RemoteTrack *webrtc.TrackRemote
}

type TrackInfoKind int

const (
	TrackInfoKindPeer TrackInfoKind = iota + 1
	TrackInfoKindStream
)

func newTrackInfo(id uuid.UUID, track *webrtc.TrackLocalStaticRTP, remoteTrack *webrtc.TrackRemote, kind TrackInfoKind) *TrackInfo {
	return &TrackInfo{
		SessionId:   id,
		Track:       track,
		RemoteTrack: remoteTrack,
		Kind:        kind,
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

func (t *TrackInfo) GetRemoteTrack() *webrtc.TrackRemote {
	return t.RemoteTrack
}
