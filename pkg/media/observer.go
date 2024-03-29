package media

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/pkg/message"
)

type msgObserver interface {
	// Use the responseId and  responseMsgNumber for your answer, so that the server knew which request you are responding to.
	OnOffer(sdp *webrtc.SessionDescription, responseId uint32, responseMsgNumber uint32)
	OnAnswer(sdp *webrtc.SessionDescription, responseId uint32, responseMsgNumber uint32)
	OnMute(mute *message.Mute)
	GetId() uuid.UUID
}
