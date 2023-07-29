package lobby

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/logging"
	"github.com/stretchr/testify/assert"
)

func testLobbyManagerSetup(t *testing.T) (*LobbyManager, *lobby) {
	t.Helper()
	logging.SetupDebugLogger()
	engine := mockRtpEngineForOffer(mockedAnswer)

	manager := NewLobbyManager(engine)
	lobby := manager.lobbies.getOrCreateLobby(uuid.New())
	return manager, lobby
}
func TestLobbyManager(t *testing.T) {
	t.Run("Access a Lobby with timeout", func(t *testing.T) {
		manager, lobby := testLobbyManagerSetup(t)
		userId := uuid.New()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // trigger cancel for time out
		_, err := manager.AccessLobby(ctx, lobby.Id, userId, mockedOffer)
		assert.ErrorIs(t, err, errLobbyRequestTimeout)
	})

	t.Run("Access a new Lobby", func(t *testing.T) {
		manager, lobby := testLobbyManagerSetup(t)
		userId := uuid.New()
		data, err := manager.AccessLobby(context.Background(), lobby.Id, userId, mockedOffer)

		assert.NoError(t, err)
		assert.Equal(t, mockedAnswer, data.Answer)
		assert.False(t, uuid.Nil == data.RtpSessionId)
		assert.False(t, uuid.Nil == data.Resource)
	})

	t.Run("Access a already started Lobby", func(t *testing.T) {
		manager, lobby := testLobbyManagerSetup(t)

		_, err := manager.AccessLobby(context.Background(), lobby.Id, uuid.New(), mockedOffer)
		assert.NoError(t, err)

		data, err := manager.AccessLobby(context.Background(), lobby.Id, uuid.New(), mockedOffer)
		assert.NoError(t, err)

		assert.Equal(t, mockedAnswer, data.Answer)
		assert.False(t, uuid.Nil == data.RtpSessionId)
		assert.False(t, uuid.Nil == data.Resource)
	})

	t.Run("Start listen to a Lobby with timeout", func(t *testing.T) {
		manager, lobby := testLobbyManagerSetup(t)
		userId := uuid.New()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // trigger cancel for time out
		_, err := manager.StartListenLobby(ctx, lobby.Id, userId)
		assert.ErrorIs(t, err, errLobbyRequestTimeout)
	})
}
