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

	entity         *LobbyEntity
	hub            *sessions.Hub
	sessions       *sessions.SessionRepository
	rtp            sessions.RtpEngine
	sessionCreator chan<- sessions.Item
	sessionGarbage chan<- sessions.Item
	lobbyGarbage   chan<- uuid.UUID
}

func newLobby(entity *LobbyEntity, rtp sessions.RtpEngine, lobbyGarbage chan<- uuid.UUID) *lobby {
	ctx, stop := context.WithCancel(context.Background())
	sessRep := sessions.NewSessionRepository()
	hub := sessions.NewHub(ctx, sessRep, entity.LiveStreamId, nil)
	sessionGarbage := make(chan sessions.Item)
	sessionCreator := make(chan sessions.Item)

	lobby := &lobby{
		Id:   entity.UUID,
		ctx:  ctx,
		stop: stop,

		hub:            hub,
		entity:         entity,
		sessions:       sessRep,
		rtp:            rtp,
		sessionGarbage: sessionGarbage,
		sessionCreator: sessionCreator,
		lobbyGarbage:   lobbyGarbage,
	}
	go func() {
		for {
			select {
			case item := <-sessionCreator:
				select {
				case <-lobby.ctx.Done():
					item.Done <- false
				default:
					ok := lobby.sessions.New(sessions.NewSession(lobby.ctx, item.UserId, lobby.hub, lobby.rtp))
					item.Done <- ok
				}
			case item := <-sessionGarbage:
				ok := lobby.sessions.DeleteByUser(item.UserId)
				item.Done <- ok
				if lobby.sessions.Len() == 0 {
					lobby.stop()
					go func() {
						lobby.lobbyGarbage <- lobby.Id
					}()
				}

			case <-lobby.ctx.Done():
				slog.Debug("stop session garbage collector")
				return
			}
		}
	}()
	return lobby
}

func (l *lobby) newSession(userId uuid.UUID) bool {
	item := sessions.NewItem(userId)
	select {
	case l.sessionCreator <- item:
		ok := <-item.Done
		return ok
	case <-l.ctx.Done():
		slog.Debug("can not adding session because lobby stopped")
		return false
	}
}

func (l *lobby) removeSession(userId uuid.UUID) bool {
	item := sessions.NewItem(userId)
	select {
	case l.sessionGarbage <- item:
		ok := <-item.Done
		return ok
	case <-l.ctx.Done():
		slog.Debug("can not remove session because lobby stopped")
		return false
	}
}

func (l *lobby) handle(cmd command) {
	if session, found := l.sessions.FindByUserId(cmd.GetUserId()); found {
		cmd.Execute(l.ctx, session)
	}
	cmd.Fail(ErrNoSession)
}
