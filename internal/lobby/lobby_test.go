package lobby

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
	"github.com/shigde/sfu/internal/logging"
	"github.com/stretchr/testify/assert"
)

func testLobbySetup(t *testing.T) (*lobby, uuid.UUID) {
	t.Helper()
	logging.SetupDebugLogger()
	entity := &LobbyEntity{
		UUID:         uuid.New(),
		LiveStreamId: uuid.New(),
		Space:        "space",
		Host:         "http://localhost:1234/federation/accounts/shig-test",
	}

	lobby := newLobby(entity, nil, make(chan<- uuid.UUID, 1))
	user := uuid.New()
	lobby.newSession(user)
	return lobby, user
}
func TestLobby_handle(t *testing.T) {
	t.Run("command successfully", func(t *testing.T) {
		lobby, user := testLobbySetup(t)
		cmd := &mockCmd{
			user: user,
			f: func(ctx context.Context, s *sessions.Session) (*resources.WebRTC, error) {
				return &resources.WebRTC{SDP: MockedAnswer}, nil
			},
			Response: make(chan *resources.WebRTC),
			Err:      make(chan error),
		}
		go lobby.handle(cmd)

		select {
		case res := <-cmd.Response:
			assert.Equal(t, MockedAnswer, res.SDP)
		case <-cmd.Err:
			t.Fatalf("no error expected")
		case <-time.After(time.Second * 3):
			t.Fatalf("run in timeout")
		}
	})

	t.Run("command error session not found", func(t *testing.T) {
		lobby, _ := testLobbySetup(t)
		cmd := &mockCmd{
			user: uuid.New(),
			Err:  make(chan error),
		}
		go lobby.handle(cmd)

		select {
		case <-cmd.Response:
			t.Fatalf("no webrtc resource expected")
		case err := <-cmd.Err:
			assert.ErrorIs(t, err, ErrNoSession)
		case <-time.After(time.Second * 3):
			t.Fatalf("run in timeout")
		}
	})

	t.Run("command with error", func(t *testing.T) {
		cmdErr := errors.New("cmd test error")
		lobby, user := testLobbySetup(t)
		cmd := &mockCmd{
			user: user,
			f: func(ctx context.Context, s *sessions.Session) (*resources.WebRTC, error) {
				return nil, cmdErr
			},
			Err: make(chan error),
		}
		go lobby.handle(cmd)

		select {
		case <-cmd.Response:
			t.Fatalf("no webrtc resource expected")
		case err := <-cmd.Err:
			assert.ErrorIs(t, err, cmdErr)
		case <-time.After(time.Second * 3):
			t.Fatalf("run in timeout")
		}
	})

	t.Run("command fails, because lobby context was done", func(t *testing.T) {
		cmdCtxErr := errors.New("cmd context done")
		lobby, user := testLobbySetup(t)
		ctx, cancel := context.WithCancel(context.Background())
		lobby.ctx = ctx
		cmd := &mockCmd{
			user: user,
			f: func(paramCtx context.Context, s *sessions.Session) (*resources.WebRTC, error) {
				select {
				case <-paramCtx.Done():
					return nil, cmdCtxErr
				default:
					return nil, nil
				}
			},
			Err:      make(chan error),
			Response: make(chan *resources.WebRTC),
		}
		cancel()
		go lobby.handle(cmd)

		select {
		case <-cmd.Response:
			t.Fatalf("no webrtc resource expected")
		case err := <-cmd.Err:
			assert.ErrorIs(t, err, cmdCtxErr)
		case <-time.After(time.Second * 3):
			t.Fatalf("run in timeout")
		}
	})
}

func TestLobby_newSession(t *testing.T) {
	t.Run("new session added", func(t *testing.T) {
		lobby, _ := testLobbySetup(t)
		user := uuid.New()
		ok := lobby.newSession(user)
		assert.True(t, ok)
		_, found := lobby.sessions.FindByUserId(user)
		assert.True(t, found)
	})

	t.Run("not add already existing session", func(t *testing.T) {
		lobby, user := testLobbySetup(t)
		ok := lobby.newSession(user)
		assert.False(t, ok)
		_, found := lobby.sessions.FindByUserId(user)
		assert.True(t, found)
	})
}

func TestLobby_removeSession(t *testing.T) {
	t.Run("delete session if exits", func(t *testing.T) {
		lobby, user := testLobbySetup(t)
		ok := lobby.removeSession(user)
		assert.True(t, ok)
		_, foundAfter := lobby.sessions.FindByUserId(user)
		assert.False(t, foundAfter)
	})

	t.Run("delete session if not exits", func(t *testing.T) {
		lobby, _ := testLobbySetup(t)
		ok := lobby.removeSession(uuid.New())
		assert.False(t, ok)
	})
}

func TestLobby_sessionGarbage(t *testing.T) {
	t.Run("delete session if exits by garbage collector", func(t *testing.T) {
		lobby, user := testLobbySetup(t)
		item := sessions.NewItem(user)

		lobby.sessionGarbage <- item
		ok := <-item.Done
		assert.True(t, ok)

		_, foundAfter := lobby.sessions.FindByUserId(user)
		assert.False(t, foundAfter)
	})

	t.Run("delete session if not exits by garbage collector", func(t *testing.T) {
		lobby, _ := testLobbySetup(t)
		item := sessions.NewItem(uuid.New())
		lobby.sessionGarbage <- item
		ok := <-item.Done
		assert.False(t, ok)
	})
}

func TestLobby_sessionSequence(t *testing.T) {
	t.Run("stop lobby when last session deleted", func(t *testing.T) {
		lobby, user1 := testLobbySetup(t)
		user2 := uuid.New()
		user3 := uuid.New()
		user4 := uuid.New()

		ok := lobby.newSession(user2)
		assert.True(t, ok)
		ok = lobby.newSession(user3)
		assert.True(t, ok)

		ok = lobby.removeSession(user2)
		assert.True(t, ok)
		ok = lobby.removeSession(user3)
		assert.True(t, ok)
		ok = lobby.removeSession(user1)
		assert.True(t, ok)

		ok = lobby.newSession(user4)
		assert.False(t, ok)
		assert.Equal(t, 0, lobby.sessions.Len())
		select {
		case <-lobby.ctx.Done():
			assert.True(t, true)
		default:
			t.Fatalf("lobby should be closed")
		}
	})
}

type mockCmd struct {
	ctx      context.Context
	user     uuid.UUID
	Response chan *resources.WebRTC
	Err      chan error
	f        func(ctx context.Context, s *sessions.Session) (*resources.WebRTC, error)
}

func (mc *mockCmd) GetUserId() uuid.UUID {
	return mc.user
}

func (mc *mockCmd) Execute(ctx context.Context, session *sessions.Session) {
	res, err := mc.f(ctx, session)
	if err != nil {
		mc.Err <- err
		return
	}
	mc.Response <- res
}

func (mc *mockCmd) Fail(err error) {
	mc.Err <- err
}
