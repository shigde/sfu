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
	reqChan    chan *lobbyRequest
}

func newLobby(id uuid.UUID, rtpEngine rtpEngine) *lobby {
	sessions := newSessionRepository()
	quitChan := make(chan struct{})
	reqChan := make(chan *lobbyRequest)
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
			switch requestType := req.data.(type) {
			case *joinData:
				l.handleJoin(req)
			case *listenData:
				l.handleListen(req)
			case *leaveData:
				l.handleLeave(req)
			default:
				slog.Error("lobby.lobby: not supported request type in Lobby", "type", requestType)
			}
		case <-l.quit:
			slog.Info("lobby.lobby: close Lobby", "id", l.Id)
			return
		}
	}
}

func (l *lobby) runRequest(req *lobbyRequest) {
	slog.Debug("lobby.lobby: runRequest", "id", l.Id)
	select {
	case l.reqChan <- req:
		slog.Debug("lobby.lobby: runRequest - requested", "id", l.Id)
	case <-l.quit:
		req.err <- errRtpSessionAlreadyClosed
		slog.Debug("lobby.lobby: runRequest - interrupted because lobby closed", "id", l.Id)
	case <-time.After(lobbyReqTimeout):
		slog.Error("lobby.lobby: runRequest - interrupted because request timeout", "id", l.Id)
	}
}

func (l *lobby) handleJoin(joinReq *lobbyRequest) {
	slog.Info("lobby.lobby: handle join", "id", l.Id, "user", joinReq.user)
	data, _ := joinReq.data.(*joinData)
	session, ok := l.sessions.FindByUserId(joinReq.user)
	if !ok {
		session = newSession(joinReq.user, l.hub, l.rtpEngine)
		l.sessions.Add(session)
	}
	offerReq := newOfferRequest(joinReq.ctx, data.offer, offerTypeReceving)

	go func() {
		slog.Info("lobby.lobby: create offer request", "id", l.Id)
		session.runOfferRequest(offerReq)
	}()
	select {
	case answer := <-offerReq.answer:
		data.response <- &joinResponse{
			answer:       answer,
			resource:     l.resourceId,
			RtpSessionId: session.Id,
		}
	case err := <-offerReq.err:
		joinReq.err <- fmt.Errorf("start session for joing: %w", err)
	case <-joinReq.ctx.Done():
		joinReq.err <- errLobbyRequestTimeout
	case <-l.quit:
		joinReq.err <- errLobbyStopped
	}
}

func (l *lobby) handleListen(req *lobbyRequest) {
	slog.Info("lobby.lobby: handle listen", "id", l.Id, "user", req.user)
	data, _ := req.data.(*listenData)
	session, ok := l.sessions.FindByUserId(req.user)
	if !ok {
		session = newSession(req.user, l.hub, l.rtpEngine)
		l.sessions.Add(session)
	}
	offerReq := newOfferRequest(req.ctx, data.offer, offerTypeSending)

	go func() {
		slog.Info("lobby.lobby: create offer request", "id", l.Id)
		session.runOfferRequest(offerReq)
	}()
	select {
	case answer := <-offerReq.answer:
		data.response <- &listenResponse{
			answer:       answer,
			RtpSessionId: session.Id,
		}
	case err := <-offerReq.err:
		req.err <- fmt.Errorf("start session for listening: %w", err)
	case <-req.ctx.Done():
		req.err <- errLobbyRequestTimeout
	case <-l.quit:
		req.err <- errLobbyStopped
	}
}

func (l *lobby) handleLeave(req *lobbyRequest) {
	slog.Info("lobby.lobby: leave", "id", l.Id, "user", req.user)
	if session, ok := l.sessions.FindByUserId(req.user); ok {
		if err := session.stop(); err != nil {
			req.err <- fmt.Errorf("stopping rtp session %s for user %s: %w", session.Id, req.user, err)
		}
		data, _ := req.data.(*leaveData)
		data.response <- l.sessions.Delete(session.Id)
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
