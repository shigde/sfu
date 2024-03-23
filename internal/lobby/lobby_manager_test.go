package lobby

import (
	"context"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/mocks"
	"github.com/shigde/sfu/internal/storage"
	"github.com/stretchr/testify/assert"
)

func testLobbyManagerSetup(t *testing.T) (*LobbyManager, uuid.UUID) {
	t.Helper()
	homeUrl, _ := url.Parse("http://localhost:1234/")
	registerToken := "federation_registration_token"
	rtp := mocks.NewRtpEngineForOffer(mocks.Answer)
	manager := NewLobbyManager(storage.NewTestStore(), rtp, homeUrl, registerToken)
	lobby, _ := testLobbySetup(t)
	manager.lobbies.lobbies[lobby.Id] = lobby
	return manager, lobby.Id
}

func TestLobbyManager_NewIngressResource(t *testing.T) {
	t.Run("get webrtc resource", func(t *testing.T) {
		manager, lobbyId := testLobbyManagerSetup(t)

		resource, err := manager.NewIngressResource(context.Background(), lobbyId, uuid.New(), mocks.Offer)
		assert.NoError(t, err)
		assert.NotNil(t, resource)
	})
}
