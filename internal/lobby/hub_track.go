package lobby

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type HubTrack interface {
	GetStreamKind() string
	GetSessionId() uuid.UUID
	GetTrack() *webrtc.TrackLocalStaticRTP
}

type TrackInfo struct {
	Kind      string
	SessionId uuid.UUID
	Track     *webrtc.TrackLocalStaticRTP
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
