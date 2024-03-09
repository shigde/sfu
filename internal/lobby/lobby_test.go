package lobby

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/logging"
)

func testStreamLobbySetup(t *testing.T) (*lobby, uuid.UUID) {
	t.Helper()
	logging.SetupDebugLogger()
	// set one session in lobby
	_ = mockRtpEngineForOffer(MockedAnswer)
	entity := &LobbyEntity{
		UUID:         uuid.New(),
		LiveStreamId: uuid.New(),
		Space:        "space",
		Host:         "http://localhost:1234/federation/accounts/shig-test",
	}

	lobby := newLobby(entity.UUID, entity)
	user := uuid.New()
	//session := sessions.NewSession(user, lobby.hub, engine, lobby.sessionQuit)
	////session.signal.messenger = newMockedMessenger(t)
	////session.ingress = mockConnection(MockedAnswer)
	////
	////session.egress = mockConnection(MockedAnswer)
	////session.signal.egress = session.egress
	//lobby.sessions.Add(session)
	return lobby, user
}
func TestStreamLobby(t *testing.T) {

	t.Run("new ingress egress", func(t *testing.T) {

	})
}
