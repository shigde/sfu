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

	lobby := newLobby(entity.UUID, entity)
	user := uuid.New()
	lobby.newSession(user, nil)
	return lobby, user
}
func TestLobby(t *testing.T) {
	t.Run("handle command successfully", func(t *testing.T) {
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
			t.Fatalf("test fails because no error expected")
		case <-time.After(time.Second * 3):
			t.Fatalf("test fails because run in timeout")
		}
	})

	t.Run("handle, command error session not found", func(t *testing.T) {
		lobby, _ := testLobbySetup(t)
		cmd := &mockCmd{
			user: uuid.New(),
			Err:  make(chan error),
		}
		go lobby.handle(cmd)

		select {
		case <-cmd.Response:
			t.Fatalf("test fails because no webrtc resource expected")
		case err := <-cmd.Err:
			assert.ErrorIs(t, err, errNoSession)
		case <-time.After(time.Second * 3):
			t.Fatalf("test fails because run in timeout")
		}
	})

	t.Run("handle command with error", func(t *testing.T) {
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
			t.Fatalf("test fails because no webrtc resource expected")
		case err := <-cmd.Err:
			assert.ErrorIs(t, err, cmdErr)
		case <-time.After(time.Second * 3):
			t.Fatalf("test fails because run in timeout")
		}
	})

	t.Run("handle command fails, because lobby context was done", func(t *testing.T) {
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
			t.Fatalf("test fails because no webrtc resource expected")
		case err := <-cmd.Err:
			assert.ErrorIs(t, err, cmdCtxErr)
		case <-time.After(time.Second * 3):
			t.Fatalf("test fails because run in timeout")
		}
	})

	t.Run("new session added", func(t *testing.T) {
		lobby, _ := testLobbySetup(t)
		user := uuid.New()
		ok := lobby.newSession(user, nil)
		assert.True(t, ok)
	})

	t.Run("new session not added", func(t *testing.T) {
		lobby, user := testLobbySetup(t)
		ok := lobby.newSession(user, nil)
		assert.False(t, ok)
	})

	t.Run("delete session if exits", func(t *testing.T) {
		lobby, user := testLobbySetup(t)
		ok := lobby.removeSession(user)
		assert.True(t, ok)
	})

	t.Run("delete session if not exits", func(t *testing.T) {
		lobby, _ := testLobbySetup(t)
		ok := lobby.removeSession(uuid.New())
		assert.False(t, ok)
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
