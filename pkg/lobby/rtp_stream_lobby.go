package lobby

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type RtpStreamLobby struct {
	Id         string
	locker     *sync.RWMutex
	sessions   map[uuid.UUID]*rtpSession
	resourceId uuid.UUID
}

func newRtpStreamLobby(id string) *RtpStreamLobby {
	s := make(map[uuid.UUID]*rtpSession)
	return &RtpStreamLobby{Id: id, resourceId: uuid.New(), sessions: s}
}

type RtpResourceData struct {
	Answer       *webrtc.SessionDescription
	Resource     uuid.UUID
	RtpSessionId uuid.UUID
}

func (l *RtpStreamLobby) Join(user uuid.UUID, offer *webrtc.SessionDescription) (*RtpResourceData, error) {
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
	return &RtpResourceData{
		Answer:       answer,
		Resource:     l.resourceId,
		RtpSessionId: session.Id,
	}, nil
}
