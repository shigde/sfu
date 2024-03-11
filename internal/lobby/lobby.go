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

// lobby, is a container for all sessions of a stream
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
	lobbyGarbage   chan<- lobbyItem
}

func newLobby(entity *LobbyEntity, rtp sessions.RtpEngine, lobbyGarbage chan<- lobbyItem) *lobby {
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
	// create and delete session should be sequentiell
	go func() {
		for {
			select {
			case item := <-sessionCreator:
				// in the meantime the lobby could be closed, check again
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
					item := newLobbyItem(lobby.Id)
					go func() {
						lobby.lobbyGarbage <- item
					}()
					<-item.Done
					// block all callers until lobby was clean up,
					// because we want to avoid that`s callers would call more than one time to crete a session
					lobby.stop()
				}
			case <-lobby.ctx.Done():
				slog.Debug("stop session sequencer")
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

// handle, run session commands on existing sessions
// the session could be already deleted, after the command was started.
// But this cse is handel by the session it selves
func (l *lobby) handle(cmd command) {
	if session, found := l.sessions.FindByUserId(cmd.GetUserId()); found {
		cmd.Execute(l.ctx, session)
	}
	cmd.Fail(ErrNoSession)
}
