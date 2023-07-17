package lobby

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

var errLobbyStopped = errors.New("error because lobby stopped")
var lobbyReqTimeout = 3 * time.Second

type rtpStreamLobby struct {
	Id         uuid.UUID
	sessions   map[uuid.UUID]*rtpSession
	rtpEngine  rtpEngine
	resourceId uuid.UUID
	quit       chan struct{}
	reqChan    chan interface{}
}

func newRtpStreamLobby(id uuid.UUID, rtpEngine rtpEngine) *rtpStreamLobby {
	sessions := make(map[uuid.UUID]*rtpSession)
	quitChan := make(chan struct{})
	reqChan := make(chan interface{})
	lobby := &rtpStreamLobby{
		Id:         id,
		resourceId: uuid.New(),
		rtpEngine:  rtpEngine,
		sessions:   sessions,
		quit:       quitChan,
		reqChan:    reqChan,
	}
	go lobby.run()
	return lobby
}

func (l *rtpStreamLobby) run() {
	slog.Info("lobby.rtpStreamLobby: run", "id", l.Id)
	for {
		select {
		case req := <-l.reqChan:
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

func (l *rtpStreamLobby) runJoin(joinReq *joinRequest) {
	slog.Debug("lobby.rtpStreamLobby: join", "id", l.Id)
	select {
	case l.reqChan <- joinReq:
		slog.Debug("lobby.rtpStreamLobby: join - join requested", "id", l.Id)
	case <-l.quit:
		joinReq.err <- errRtpSessionAlreadyClosed
		slog.Debug("lobby.rtpStreamLobby: join - interrupted because lobby closed", "id", l.Id)
	case <-time.After(lobbyReqTimeout):
		slog.Error("lobby.rtpStreamLobby: join - interrupted because request timeout", "id", l.Id)
	}
}

func (l *rtpStreamLobby) handleJoin(joinReq *joinRequest) {
	slog.Info("lobby.rtpStreamLobby: handle join", "id", l.Id, "user", joinReq.user)
	session, ok := l.sessions[joinReq.user]
	if !ok {
		session = newRtpSession(joinReq.user, l.rtpEngine)
		l.sessions[joinReq.user] = session
	}
	offerReq := newOfferRequest(joinReq.ctx, joinReq.offer)

	go func() {
		// @TODO: We have to catch the case join gets canceled but session was not already finish offer.
		// In this case we become a ghosts session!
		// Pass the context to the session is the best way to do this.
		// @TODO Here we hav a race condition. I will change this in a way that an offer will not be faster as an run session.
		// answer, err = session.offer(joinReq.offer)
		slog.Info("lobby.rtpStreamLobby: create offer request", "id", l.Id)
		session.runOffer(offerReq)
	}()
	select {
	case answer := <-offerReq.answer:
		joinReq.response <- &joinResponse{
			answer:       answer,
			resource:     l.resourceId,
			RtpSessionId: session.Id,
		}
	case err := <-offerReq.err:
		joinReq.err <- fmt.Errorf("start session for joiing: %w", err)
	case <-joinReq.ctx.Done():
		joinReq.err <- errLobbyRequestTimeout
	case <-l.quit:
		joinReq.err <- errLobbyStopped
	}
}

func (l *rtpStreamLobby) handleLeave(req *leaveRequest) {
	slog.Info("lobby.rtpStreamLobby: leave", "id", l.Id, "user", req.user)
	if session, ok := l.sessions[req.user]; ok {
		if err := session.stop(); err != nil {
			req.err <- fmt.Errorf("stopping rtp session %s for user %s: %w", session.Id, req.user, err)
		}
		delete(l.sessions, req.user)
		req.response <- true
		return
	}
	req.err <- fmt.Errorf("no session existing for user %s", req.user)
}

func (l *rtpStreamLobby) stop() {
	slog.Info("lobby.rtpStreamLobby: stop", "id", l.Id)
	select {
	case <-l.quit:
		slog.Warn("lobby.rtpStreamLobby: the Lobby was already closed", "id", l.Id)
	default:
		close(l.quit)
		slog.Info("lobby.rtpStreamLobby: stopped was triggered", "id", l.Id)
	}
}

type joinRequest struct {
	user     uuid.UUID
	offer    *webrtc.SessionDescription
	response chan *joinResponse
	err      chan error
	ctx      context.Context
}

func newJoinRequest(ctx context.Context, user uuid.UUID, offer *webrtc.SessionDescription) *joinRequest {
	errChan := make(chan error)
	resChan := make(chan *joinResponse)

	return &joinRequest{
		offer:    offer,
		user:     user,
		err:      errChan,
		response: resChan,
		ctx:      ctx,
	}
}

type joinResponse struct {
	answer       *webrtc.SessionDescription
	resource     uuid.UUID
	RtpSessionId uuid.UUID
}

type leaveRequest struct {
	user     uuid.UUID
	response chan bool
	err      chan error
	ctx      context.Context
}
