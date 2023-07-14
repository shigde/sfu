package lobby

import (
	"errors"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

var errRtpSessionAlreadyClosed = errors.New("the rtp session was already closed")

type rtpSession struct {
	Id        uuid.UUID
	user      uuid.UUID
	streams   *rtpStreamRepository
	rtpEngine rtpEngine
	quit      chan struct{}
}

func newRtpSession(user uuid.UUID, e rtpEngine) *rtpSession {
	repo := newRtpStreamRepository()
	q := make(chan struct{})
	session := &rtpSession{
		Id:        uuid.New(),
		user:      user,
		streams:   repo,
		rtpEngine: e,
		quit:      q,
	}
	go session.run()
	return session
}

func (s *rtpSession) run() {
	slog.Info("lobby.rtpSession: run", "id", s.Id, "user", s.user)
	for {
		select {
		case <-s.quit:
			// @TODO Take care that's every stream is closed!
			slog.Info("lobby.rtpSession: close rtp session", "id", s.Id, "user", s.user)
			return
		}
	}
}

func (s *rtpSession) offer(_ *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	slog.Info("lobby.rtpSession: offer", "id", s.Id, "user", s.user)
	return nil, nil
}

func (s *rtpSession) stop() error {
	slog.Info("lobby.rtpSession: stop", "id", s.Id, "user", s.user)
	select {
	case <-s.quit:
		slog.Error("lobby.rtpSession: the rtp session was already closed", "id", s.Id, "user", s.user)
		return errRtpSessionAlreadyClosed
	default:
		close(s.quit)
	}
	return nil
}
