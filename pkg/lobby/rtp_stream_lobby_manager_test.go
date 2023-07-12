package lobby

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
)

func testLobbyManagerSetup(t *testing.T) (*RtpStreamLobbyManager, *rtpStreamLobby) {
	t.Helper()
	manager := NewLobbyManager()
	lobby := manager.lobbies.getOrCreateLobby(uuid.New())
	return manager, lobby
}
func TestLobbyManager(t *testing.T) {
	t.Run("Access a Lobby with timeout", func(t *testing.T) {
		manager, lobby := testLobbyManagerSetup(t)
		var offer *webrtc.SessionDescription
		userId := uuid.New()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // trigger cancel for time out
		_, err := manager.AccessLobby(ctx, lobby.Id, userId, offer)
		assert.Error(t, err, errLobbyRequestTimeout)
	})

	t.Run("Access a Lobby with timeout", func(t *testing.T) {
		manager, lobby := testLobbyManagerSetup(t)
		var offer *webrtc.SessionDescription
		userId := uuid.New()
		data, err := manager.AccessLobby(context.Background(), lobby.Id, userId, offer)

		assert.NoError(t, err)
		assert.Nil(t, data.Answer)
		assert.False(t, uuid.Nil == data.RtpSessionId)
		assert.False(t, uuid.Nil == data.Resource)
	})
}
