package lobby

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type rtpSession struct {
	Id      uuid.UUID
	User    uuid.UUID
	streams *rtpStreamRepository
}

func newRtpSession(user uuid.UUID) *rtpSession {
	repo := newRtpStreamRepository()
	return &rtpSession{
		Id:      uuid.New(),
		User:    user,
		streams: repo,
	}
}

func (s *rtpSession) Join(_ *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	return nil, nil
}
