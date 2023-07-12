package lobby

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

type rtpStreamLobby struct {
	Id         uuid.UUID
	sessions   map[uuid.UUID]*rtpSession
	resourceId uuid.UUID
	quit       chan struct{}
	request    chan interface{}
}

func newRtpStreamLobby(id uuid.UUID) *rtpStreamLobby {
	s := make(map[uuid.UUID]*rtpSession)
	q := make(chan struct{})
	r := make(chan interface{})
	lobby := &rtpStreamLobby{
		Id:         id,
		resourceId: uuid.New(),
		sessions:   s,
		quit:       q,
		request:    r,
	}
	go lobby.run()
	return lobby
}

func (l *rtpStreamLobby) run() {
	slog.Info("lobby.rtpStreamLobby: run rtp session", "id", l.Id)
	for {
		select {
		case req := <-l.request:
			switch v := req.(type) {
			case *joinRequest:
				l.handleJoin(v)
			case *leaveRequest:
				l.handleLeave(v)
			default:
				slog.Error("lobby.rtpStreamLobby: not supported request type in Lobby", "type", v)
			}
		case <-l.quit:
			slog.Info("lobby.rtpStreamLobby: close Lobby", "id", l.Id)
			return
		}
	}
}

func (l *rtpStreamLobby) handleJoin(req *joinRequest) {
	session, ok := l.sessions[req.user]
	if !ok {
		session = newRtpSession(req.user)
		l.sessions[req.user] = session
	}

	done := make(chan struct{})
	var answer *webrtc.SessionDescription
	var err error

	go func() {
		// @TODO: We have to catch the case join gets canceled but session was not already finish offer.
		// In this case we become a ghosts session!
		// Pass the context to the session is the best way to do this.
		answer, err = session.offer(req.offer)
		done <- struct{}{}
	}()
	select {
	case <-done:
		if err != nil {
			req.error <- fmt.Errorf("joining rtp session: %w", err)
			return
		}
		req.response <- &joinResponse{
			answer:       answer,
			resource:     l.resourceId,
			RtpSessionId: session.Id,
		}
	case <-req.ctx.Done():
		req.error <- errLobbyRequestTimeout
	}
}

func (l *rtpStreamLobby) handleLeave(req *leaveRequest) {
	if session, ok := l.sessions[req.user]; ok {
		if err := session.stop(); err != nil {
			req.error <- fmt.Errorf("stopping rtp session %s for user %s: %w", session.Id, req.user, err)
		}
		delete(l.sessions, req.user)
		req.response <- true
		return
	}
}

func (l *rtpStreamLobby) stop() {
	select {
	case <-l.quit:
		slog.Warn("lobby.rtpStreamLobby: the Lobby was already closed", "id", l.Id)
	default:
		close(l.quit)
	}
}

type joinRequest struct {
	user     uuid.UUID
	offer    *webrtc.SessionDescription
	response chan *joinResponse
	error    chan error
	ctx      context.Context
}

type joinResponse struct {
	answer       *webrtc.SessionDescription
	resource     uuid.UUID
	RtpSessionId uuid.UUID
}

type leaveRequest struct {
	user     uuid.UUID
	response chan bool
	error    chan error
	ctx      context.Context
}
