package lobby

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/sessions"
	"golang.org/x/exp/slog"
)

type command interface {
	GetUserId() uuid.UUID
	Execute(ctx context.Context, session *sessions.Session)
	Fail(err error)
}

var (
	ErrNoSession            = errors.New("no session exists")
	ErrSessionAlreadyExists = errors.New("session already exists")
)

type lobby struct {
	Id   uuid.UUID
	ctx  context.Context
	stop context.CancelFunc

	entity                  *LobbyEntity
	hub                     *sessions.Hub
	sessions                *sessions.SessionRepository
	sessionGarbageCollector chan<- uuid.UUID
}

func newLobby(entity *LobbyEntity, garbage chan<- uuid.UUID) *lobby {
	ctx, stop := context.WithCancel(context.Background())
	sessRep := sessions.NewSessionRepository()
	hub := sessions.NewHub(ctx, sessRep, entity.LiveStreamId, nil)
	sessionGarbageCollector := make(chan uuid.UUID)
	lobby := &lobby{
		Id:   entity.UUID,
		ctx:  ctx,
		stop: stop,

		hub:                     hub,
		entity:                  entity,
		sessions:                sessRep,
		sessionGarbageCollector: sessionGarbageCollector,
	}
	go func() {
		for {
			select {
			case userId := <-sessionGarbageCollector:
				if ok := lobby.sessions.DeleteByUser(userId); !ok {
					slog.Warn("session could not delete", "session id", userId)
				}
			case <-lobby.ctx.Done():
				slog.Debug("stop session garbage collector")
				return
			}
		}
	}()
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
	cmd.Fail(ErrNoSession)
}
