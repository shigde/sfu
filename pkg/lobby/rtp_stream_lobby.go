package lobby

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type rtpStreamLobby struct {
	Id         uuid.UUID
	locker     *sync.RWMutex
	sessions   map[uuid.UUID]*rtpSession
	resourceId uuid.UUID
	quit       chan struct{}
	active     bool
}

func newRtpStreamLobby(id uuid.UUID) *rtpStreamLobby {
	s := make(map[uuid.UUID]*rtpSession)
	return &rtpStreamLobby{
		Id:         id,
		locker:     &sync.RWMutex{},
		resourceId: uuid.New(),
		sessions:   s,
		active:     false,
	}
}

type rtpResourceData struct {
	Answer       *webrtc.SessionDescription
	Resource     uuid.UUID
	RtpSessionId uuid.UUID
}

func (l *rtpStreamLobby) Join(user uuid.UUID, offer *webrtc.SessionDescription) (*rtpResourceData, error) {
	l.locker.Lock()
	defer l.locker.Unlock()

	session, ok := l.sessions[user]
	if !ok {
		session = newRtpSession(user)
		l.sessions[user] = session
	}
	answer, err := session.Join(offer)
	if err != nil {
		return nil, fmt.Errorf("joining rtp session: %w", err)
	}

	return &rtpResourceData{
		Answer:       answer,
		Resource:     l.resourceId,
		RtpSessionId: session.Id,
	}, nil
}

func (l *rtpStreamLobby) IsActive() bool {
	return l.quit != nil
}

func (l *rtpStreamLobby) Run(quit chan struct{}) {
	l.quit = quit
}
