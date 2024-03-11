package lobby

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type command interface {
	GetUserId() uuid.UUID
	Execute(ctx context.Context, session *sessions.Session)
	Fail(err error)
}

var (
	errNoSession            = errors.New("no session exists")
	ErrSessionAlreadyExists = errors.New("session already exists")
	errLobbyStopped         = errors.New("error because lobby stopped")
	errLobbyRequestTimeout  = errors.New("lobby request timeout error")
	lobbyReqTimeout         = 10 * time.Second
)

type lobby struct {
	Id   uuid.UUID
	ctx  context.Context
	stop context.CancelFunc

	entity   *LobbyEntity
	sessions *sessions.SessionRepository
	hub      *sessions.Hub
}

func newLobby(id uuid.UUID, entity *LobbyEntity) *lobby {
	ctx, stop := context.WithCancel(context.Background())
	sessRep := sessions.NewSessionRepository()
	hub := sessions.NewHub(ctx, sessRep, entity.LiveStreamId, nil)
	lobby := &lobby{
		Id:   id,
		ctx:  ctx,
		stop: stop,

		sessions: sessRep,
		hub:      hub,
		entity:   entity,
	}
	return lobby
}

func (l *lobby) newSession(userId uuid.UUID, rtp sessions.RtpEngine) bool {
	return l.sessions.New(sessions.NewSession(l.ctx, userId, l.hub, rtp))
}

func (l *lobby) removeSession(userId uuid.UUID) bool {
	return l.sessions.DeleteByUser(userId)
}

func (l *lobby) handle(cmd command) {
	if session, found := l.sessions.FindByUserId(cmd.GetUserId()); found {
		cmd.Execute(l.ctx, session)
	}
	cmd.Fail(errNoSession)
}
