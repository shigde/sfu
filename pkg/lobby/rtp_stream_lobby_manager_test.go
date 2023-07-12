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
	t.Run("Access a Lobby", func(t *testing.T) {
		manager, lobby := testLobbyManagerSetup(t)
		var offer *webrtc.SessionDescription
		userId := uuid.New()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		data, err := manager.AccessLobby(ctx, lobby.Id, userId, offer)

		assert.NotNil(t, data)
		assert.NoError(t, err)
	})
}
