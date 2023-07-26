package lobby

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

var errLobbyStopped = errors.New("error because lobby stopped")
var lobbyReqTimeout = 3 * time.Second

type lobby struct {
	Id         uuid.UUID
	sessions   *sessionRepository
	hub        *hub
	rtpEngine  rtpEngine
	resourceId uuid.UUID
	quit       chan struct{}
	reqChan    chan interface{}
}

func newLobby(id uuid.UUID, rtpEngine rtpEngine) *lobby {
	sessions := newSessionRepository()
	quitChan := make(chan struct{})
	reqChan := make(chan interface{})
	hub := newHub(sessions)
	lobby := &lobby{
		Id:         id,
		resourceId: uuid.New(),
		rtpEngine:  rtpEngine,
		sessions:   sessions,
		hub:        hub,
		quit:       quitChan,
		reqChan:    reqChan,
	}
	go lobby.run()
	return lobby
}

func (l *lobby) run() {
	slog.Info("lobby.lobby: run", "id", l.Id)
	for {
		select {
		case req := <-l.reqChan:
			switch requestType := req.(type) {
			case *joinRequest:
				l.handleJoin(requestType)
			case *leaveRequest:
				l.handleLeave(requestType)
			default:
				slog.Error("lobby.lobby: not supported request type in Lobby", "type", requestType)
			}
		case <-l.quit:
			slog.Info("lobby.lobby: close Lobby", "id", l.Id)
			return
		}
	}
}

func (l *lobby) runJoin(joinReq *joinRequest) {
	slog.Debug("lobby.lobby: join", "id", l.Id)
	select {
	case l.reqChan <- joinReq:
		slog.Debug("lobby.lobby: join - join requested", "id", l.Id)
	case <-l.quit:
		joinReq.err <- errRtpSessionAlreadyClosed
		slog.Debug("lobby.lobby: join - interrupted because lobby closed", "id", l.Id)
	case <-time.After(lobbyReqTimeout):
		slog.Error("lobby.lobby: join - interrupted because request timeout", "id", l.Id)
	}
}

func (l *lobby) handleJoin(joinReq *joinRequest) {
	slog.Info("lobby.lobby: handle join", "id", l.Id, "user", joinReq.user)
	session, ok := l.sessions.FindByUserId(joinReq.user)
	if !ok {
		session = newSession(joinReq.user, l.hub.onTrack, l.rtpEngine)
		l.sessions.Add(session)
	}
	offerReq := newOfferRequest(joinReq.ctx, joinReq.offer)

	go func() {
		slog.Info("lobby.lobby: create offer request", "id", l.Id)
		session.runOfferRequest(offerReq)
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

func (l *lobby) handleLeave(req *leaveRequest) {
	slog.Info("lobby.lobby: leave", "id", l.Id, "user", req.user)
	if session, ok := l.sessions.FindByUserId(req.user); ok {
		if err := session.stop(); err != nil {
			req.err <- fmt.Errorf("stopping rtp session %s for user %s: %w", session.Id, req.user, err)
		}
		req.response <- l.sessions.Delete(session.Id)
		return
	}
	req.err <- fmt.Errorf("no session existing for user %s", req.user)
}

func (l *lobby) stop() {
	slog.Info("lobby.lobby: stop", "id", l.Id)
	select {
	case <-l.quit:
		slog.Warn("lobby.lobby: the Lobby was already closed", "id", l.Id)
	default:
		close(l.quit)
		slog.Info("lobby.lobby: stopped was triggered", "id", l.Id)
	}
}
