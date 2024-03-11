package lobby

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/storage"
)

func testLobbyManagerSetup(t *testing.T) (*LobbyManager, uuid.UUID) {
	t.Helper()
	homeUrl, _ := url.Parse("http://localhost:1234/")
	registerToken := "federation_registration_token"
	rtp := newRtpEngineMock()
	manager := NewLobbyManager(storage.NewTestStore(), rtp, homeUrl, registerToken)
	lobbyId := uuid.New()
	return manager, lobbyId
}

func TestLobbyManager_NewIngressResource(t *testing.T) {
	t.Run("command successfully", func(t *testing.T) {
		testLobbyManagerSetup(t)
	})

}
