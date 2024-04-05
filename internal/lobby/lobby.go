package lobby

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/instances"
	"github.com/shigde/sfu/internal/lobby/sessions"
	"golang.org/x/exp/slog"
)

type command interface {
	GetUserId() uuid.UUID
	Execute(session *sessions.Session)
	SetError(err error)
}

var (
	ErrNoSession            = errors.New("no session exists")
	ErrSessionAlreadyExists = errors.New("session already exists")
	ErrLobbyClosed          = errors.New("lobby already closed")
)

// lobby, is a container for all sessions of a stream
type lobby struct {
	Id   uuid.UUID
	ctx  context.Context
	stop context.CancelFunc

	entity   *LobbyEntity
	hub      *sessions.Hub
	sessions *sessions.SessionRepository
	rtp      sessions.RtpEngine

	sessionCreator chan<- sessions.Item
	sessionGarbage chan<- sessions.Item
	lobbyGarbage   chan<- lobbyItem
	cmdRunner      chan<- command

	connector *instances.Connector
}

func newLobby(entity *LobbyEntity, rtp sessions.RtpEngine, homeActorIri *url.URL, registerToken string, lobbyGarbage chan<- lobbyItem) *lobby {
	ctx, stop := context.WithCancel(context.Background())
	sessRep := sessions.NewSessionRepository()
	hub := sessions.NewHub(ctx, sessRep, entity.LiveStreamId, nil)
	hostActorIri, _ := url.Parse(entity.Host)

	garbage := make(chan sessions.Item)
	creator := make(chan sessions.Item)
	runner := make(chan command)
	connector := instances.NewConnector(ctx, *homeActorIri, *hostActorIri, entity.Space, entity.LiveStreamId.String(), registerToken)

	lobObj := &lobby{
		Id:   entity.UUID,
		ctx:  ctx,
		stop: stop,

		hub:      hub,
		entity:   entity,
		sessions: sessRep,
		rtp:      rtp,

		sessionGarbage: garbage,
		sessionCreator: creator,
		lobbyGarbage:   lobbyGarbage,
		cmdRunner:      runner,

		connector: connector,
	}
	// session handling should be sequentiell to avoid races conditions in whole group state
	go func(l *lobby, sessionCreator <-chan sessions.Item, sessionGarbage chan sessions.Item, cmdRunner <-chan command) {
		for {
			select {
			case item := <-sessionCreator:
				// in the meantime the lobby could be closed, check again
				select {
				case <-l.ctx.Done():
					item.Done <- false
				default:
					session := sessions.NewSession(l.ctx, item.UserId, l.hub, l.rtp, item.SessionType, sessionGarbage)
					ok := l.sessions.New(session)
					item.Done <- ok
				}
			case item := <-sessionGarbage:
				ok := l.sessions.DeleteByUser(item.UserId)
				item.Done <- ok
				if l.sessions.LenUserSession() == 0 {
					item := newLobbyItem(l.Id)
					go func() {
						l.lobbyGarbage <- item
					}()
					<-item.Done
					// block all callers until lobby was clean up,
					// because we want to avoid that`s callers would call more than one time to crete a session
					l.stop()
				}
			case cmd := <-cmdRunner:
				// in the meantime the lobby could be closed, check again
				select {
				case <-l.ctx.Done():
					cmd.SetError(ErrLobbyClosed)
				default:
					l.handle(cmd)
				}
			// wenn der getriggert wird können die anderen überlaufen :-(
			case <-l.ctx.Done():
				slog.Debug("stop session sequencer")
				return
			}
		}
	}(lobObj, creator, garbage, runner)

	if !connector.IsThisInstanceLiveSteamHost() {
		go func(l *lobby) {
			l.connectToLiveStreamHostInstance()
		}(lobObj)
	}

	return lobObj
}

func (l *lobby) newSession(userId uuid.UUID, sType sessions.SessionType) bool {
	item := sessions.NewItem(userId)
	item.SessionType = sType
	select {
	case l.sessionCreator <- item:
		ok := <-item.Done
		return ok
	case <-l.ctx.Done():
		slog.Debug("can not adding session because lobby stopped")
		return false
	case <-time.After(1 * time.Second):
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
	case <-time.After(10 * time.Second):
		return false
	}
}

func (l *lobby) runCommand(cmd command) {
	select {
	case l.cmdRunner <- cmd:
	case <-l.ctx.Done():
		cmd.SetError(ErrLobbyClosed)
	}
}

// handle, run session commands on existing sessions
func (l *lobby) handle(cmd command) {
	if session, found := l.sessions.FindByUserId(cmd.GetUserId()); found {
		cmd.Execute(session)
	}
	cmd.SetError(ErrNoSession)
}

func (l *lobby) connectToLiveStreamHostInstance() {
	if ok := l.newSession(l.connector.GetInstanceId(), sessions.InstanceSession); !ok {
		slog.Error("no session found for instance connection")
		return
	}

	cmdIngress, err := l.connector.BuildIngress()
	if err != nil {
		slog.Error("build Ingress connection", "err", err)
		return
	}
	l.runCommand(cmdIngress)
	if err = cmdIngress.WaitForDone(); err != nil {
		slog.Error("run ingress build connection cmd", "err", err)
		return
	}

	cmdEgress, err := l.connector.BuildEgress()
	if err != nil {
		slog.Error("build Egress connection", "err", err)
		return
	}
	l.runCommand(cmdEgress)
	if err = cmdEgress.WaitForDone(); err != nil {
		slog.Error("run egress build connection cmd", "err", err)
		return
	}
}
