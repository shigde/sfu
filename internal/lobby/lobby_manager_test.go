package lobby

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/mocks"
	"github.com/shigde/sfu/internal/storage"
	"github.com/stretchr/testify/assert"
)

func testLobbyManagerSetup(t *testing.T) (*LobbyManager, uuid.UUID, *mocks.RtpEngineMock) {
	t.Helper()

	homeUrl, _ := url.Parse("http://localhost:1234/")
	registerToken := "federation_registration_token"
	rtp := mocks.NewRtpEngineForOffer(mocks.Answer)
	store := storage.NewTestStore()
	_ = store.GetDatabase().AutoMigrate(&LobbyEntity{})
	lobbyId := uuid.New()
	liveStreamId := uuid.New()

	entity := &LobbyEntity{
		UUID:         lobbyId,
		LiveStreamId: liveStreamId,
		Space:        "space",
		Host:         fmt.Sprintf("%s/federation/accounts/shig-test", homeUrl.Host),
	}
	store.GetDatabase().Create(entity)
	manager := NewLobbyManager(store, rtp, homeUrl, registerToken)

	return manager, lobbyId, rtp
}

func TestLobbyManager_NewIngressResource(t *testing.T) {
	t.Run("get ingress resource", func(t *testing.T) {
		manager, lobbyId, _ := testLobbyManagerSetup(t)
		resource, err := manager.NewIngressResource(context.Background(), lobbyId, uuid.New(), mocks.Offer)
		assert.NoError(t, err)
		assert.Equal(t, mocks.Answer, resource.SDP)
	})
}
