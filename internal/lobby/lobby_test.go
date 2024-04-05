package lobby

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/commands"
	"github.com/shigde/sfu/internal/lobby/mocks"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
	"github.com/shigde/sfu/internal/logging"
	"github.com/stretchr/testify/assert"
)

func testLobbySetup(t *testing.T) (*lobby, uuid.UUID) {
	t.Helper()
	logging.SetupDebugLogger()
	hostActorIri, _ := url.Parse("http://localhost:1234/federation/accounts/shig-test")
	homeActorIri, _ := url.Parse("http://localhost:1234/federation/accounts/shig-test")

	entity := &LobbyEntity{
		UUID:         uuid.New(),
		LiveStreamId: uuid.New(),
		Space:        "space",
		Host:         hostActorIri.String(),
	}

	lobby := newLobby(entity, nil, homeActorIri, "token", make(chan<- lobbyItem, 1))
	user := uuid.New()
	lobby.newSession(user, sessions.UserSession)
	return lobby, user
}
func TestLobby_handle(t *testing.T) {
	t.Run("command successfully", func(t *testing.T) {
		lobby, user := testLobbySetup(t)
		cmd := newCmdMock(context.Background(), user)
		cmd.f = func(ctx context.Context, s *sessions.Session) (*resources.WebRTC, error) {
			return &resources.WebRTC{SDP: mocks.Answer}, nil
		}
		lobby.runCommand(cmd)

		select {
		case <-cmd.Done():
			assert.Equal(t, mocks.Answer, cmd.Response.SDP)
			assert.NoError(t, cmd.Err)
		case <-time.After(time.Second * 3):
			t.Fatalf("run in timeout")
		}
	})

	t.Run("command error session not found", func(t *testing.T) {
		lobby, _ := testLobbySetup(t)
		cmd := newCmdMock(context.Background(), uuid.New())
		lobby.runCommand(cmd)

		select {
		case <-cmd.Done():
			assert.Nil(t, cmd.Response)
			assert.ErrorIs(t, cmd.Err, ErrNoSession)
		case <-time.After(time.Second * 3):
			t.Fatalf("run in timeout")
		}
	})

	t.Run("command with error", func(t *testing.T) {
		cmdErr := errors.New("cmd test error")
		lobby, user := testLobbySetup(t)
		cmd := newCmdMock(context.Background(), user)
		cmd.f = func(ctx context.Context, s *sessions.Session) (*resources.WebRTC, error) {
			return nil, cmdErr
		}
		lobby.runCommand(cmd)

		select {
		case <-cmd.Done():
			assert.Nil(t, cmd.Response)
			assert.ErrorIs(t, cmd.Err, cmdErr)
		case <-time.After(time.Second * 3):
			t.Fatalf("run in timeout")
		}
	})

	t.Run("command fails, because lobby context was done", func(t *testing.T) {
		lobby, user := testLobbySetup(t)
		cmd := newCmdMock(context.Background(), user)
		cmd.f = func(ctx context.Context, s *sessions.Session) (*resources.WebRTC, error) {
			return nil, nil
		}
		lobby.stop()
		lobby.runCommand(cmd)

		select {
		case <-cmd.Done():
			assert.Nil(t, cmd.Response)
			assert.ErrorIs(t, cmd.Err, ErrLobbyClosed)
		case <-time.After(time.Second * 3):
			t.Fatalf("run in timeout")
		}
	})
}

func TestLobby_newSession(t *testing.T) {
	t.Run("new session added", func(t *testing.T) {
		lobby, _ := testLobbySetup(t)
		user := uuid.New()
		ok := lobby.newSession(user, sessions.UserSession)
		assert.True(t, ok)
		_, found := lobby.sessions.FindByUserId(user)
		assert.True(t, found)
	})

	t.Run("not add already existing session", func(t *testing.T) {
		lobby, user := testLobbySetup(t)
		ok := lobby.newSession(user, sessions.UserSession)
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
		lobbyGarbage := make(chan lobbyItem)
		lobby.lobbyGarbage = lobbyGarbage

		user2 := uuid.New()
		user3 := uuid.New()
		user4 := uuid.New()

		ok := lobby.newSession(user2, sessions.UserSession)
		assert.True(t, ok)
		ok = lobby.newSession(user3, sessions.UserSession)
		assert.True(t, ok)

		ok = lobby.removeSession(user2)
		assert.True(t, ok)
		ok = lobby.removeSession(user3)
		assert.True(t, ok)
		ok = lobby.removeSession(user1)
		assert.True(t, ok)

		// wait for delete trigger
		item := <-lobbyGarbage
		item.Done <- true

		// we can not add a new session in a closed lobby
		ok = lobby.newSession(user4, sessions.UserSession)
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
	*commands.Command
	Response *resources.WebRTC
	f        func(ctx context.Context, s *sessions.Session) (*resources.WebRTC, error)
}

func newCmdMock(ctx context.Context, user uuid.UUID) *mockCmd {
	return &mockCmd{
		Command:  commands.NewCommand(ctx, user),
		Response: nil,
	}
}

func (mc *mockCmd) Execute(session *sessions.Session) {
	res, err := mc.f(mc.ParentCtx, session)
	if err != nil {
		mc.SetError(err)
		return
	}
	mc.Response = res
	mc.SetDone()
}
