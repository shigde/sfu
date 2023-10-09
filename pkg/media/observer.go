package media

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type msgObserver interface {
	OnOffer(sdp *webrtc.SessionDescription, number uint32)
	OnAnswer(sdp *webrtc.SessionDescription, number uint32)
	GetId() uuid.UUID
}
