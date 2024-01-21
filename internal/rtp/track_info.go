package rtp

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type TrackInfo struct {
	TrackSdpInfo
	Track *webrtc.TrackLocalStaticRTP
}

func newTrackInfo(track *webrtc.TrackLocalStaticRTP, sdpInfo TrackSdpInfo) *TrackInfo {
	return &TrackInfo{
		Track:        track,
		TrackSdpInfo: sdpInfo,
	}
}

func (t *TrackInfo) GetId() uuid.UUID {
	return t.Id
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

func (t *TrackInfo) GetMute() bool {
	return t.Mute
}
func (t *TrackInfo) GetIngressMid() string {
	return t.IngressMid
}
func (t *TrackInfo) GetEgressMid() string {
	return t.EgressMid
}

func (t *TrackInfo) SetMute(mute bool) {
	t.Mute = mute
}
