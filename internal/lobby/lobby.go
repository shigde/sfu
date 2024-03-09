package lobby

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/sessions"
	"golang.org/x/exp/slog"
)

type command interface {
	GetSessionId() uuid.UUID
	Execute(session *sessions.Session)
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

	cmd chan command
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
	go lobby.run()
	return lobby
}

func (l *lobby) run() {
	slog.Info("lobby.lobby: run", "lobbyId", l.Id)
	for {
		select {
		// case cmd := <-l.cmd:

		case <-l.ctx.Done():
			slog.Info("lobby.lobby: close lobby", "lobbyId", l.Id)
			return
		}
	}
}
