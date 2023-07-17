package lobby

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

var errLobbyStopped = errors.New("error because lobby stopped")

type rtpStreamLobby struct {
	Id         uuid.UUID
	sessions   map[uuid.UUID]*rtpSession
	engine     rtpEngine
	resourceId uuid.UUID
	quit       chan struct{}
	request    chan interface{}
}

func newRtpStreamLobby(id uuid.UUID, e rtpEngine) *rtpStreamLobby {
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
	slog.Info("lobby.rtpStreamLobby: run", "id", l.Id)
	for {
		select {
		case req := <-l.request:
			switch requestType := req.(type) {
			case *joinRequest:
				l.handleJoin(requestType)
			case *leaveRequest:
				l.handleLeave(requestType)
			default:
				slog.Error("lobby.rtpStreamLobby: not supported request type in Lobby", "type", requestType)
			}
		case <-l.quit:
			slog.Info("lobby.rtpStreamLobby: close Lobby", "id", l.Id)
			return
		}
	}
}

func (l *rtpStreamLobby) handleJoin(joinReq *joinRequest) {
	slog.Info("lobby.rtpStreamLobby: join", "id", l.Id, "user", joinReq.user)
	session, ok := l.sessions[joinReq.user]
	if !ok {
		session = newRtpSession(joinReq.user, l.engine)
		l.sessions[joinReq.user] = session
	}
	var cancel context.CancelFunc
	joinReq.ctx, cancel = context.WithCancel(joinReq.ctx)
	defer cancel()

	done := make(chan struct{})
	var answer *webrtc.SessionDescription
	var err error

	go func() {
		// @TODO: We have to catch the case join gets canceled but session was not already finish offer.
		// In this case we become a ghosts session!
		// Pass the context to the session is the best way to do this.
		// @TODO Here we hav a race condition. I will change this in a way that an offer will not be faster as an run session.
		// answer, err = session.offer(joinReq.offer)
		slog.Info("lobby.rtpStreamLobby: create offer request", "id", l.Id)
		defer cancel()
		offerReq := newOfferRequest(joinReq.ctx, joinReq.offer)

		select {
		case session.offerChan <- offerReq:
		case <-offerReq.ctx.Done():
			slog.Warn("lobby.rtpStreamLobby: offer interrupted before creating connection", "id", l.Id)
		}

		done <- struct{}{}
	}()
	select {
	case <-done:
		if err != nil {
			joinReq.error <- fmt.Errorf("joining rtp session: %w", err)
			return
		}
		joinReq.response <- &joinResponse{
			answer:       answer,
			resource:     l.resourceId,
			RtpSessionId: session.Id,
		}
	case <-joinReq.ctx.Done():
		joinReq.error <- errLobbyRequestTimeout
	case <-l.quit:
		joinReq.error <- errLobbyStopped

	}
}

func (l *rtpStreamLobby) handleLeave(req *leaveRequest) {
	slog.Info("lobby.rtpStreamLobby: leave", "id", l.Id, "user", req.user)
	if session, ok := l.sessions[req.user]; ok {
		if err := session.stop(); err != nil {
			req.error <- fmt.Errorf("stopping rtp session %s for user %s: %w", session.Id, req.user, err)
		}
		delete(l.sessions, req.user)
		req.response <- true
		return
	}
	req.error <- fmt.Errorf("no session existing for user %s", req.user)
}

func (l *rtpStreamLobby) stop() {
	slog.Info("lobby.rtpStreamLobby: stop", "id", l.Id)
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
