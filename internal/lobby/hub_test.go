package lobby

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func testHubSetup(t *testing.T) *hub {
	t.Helper()
	sessions := newSessionRepository()
	engine := newRtpEngineMock()
	hub := newHub(sessions)
	s1 := newSession(uuid.New(), hub, engine, onQuitSessionInternallyStub)
	s2 := newSession(uuid.New(), hub, engine, onQuitSessionInternallyStub)
	sessions.Add(s1)
	sessions.Add(s2)
	return hub
}
func testHubSessionSetup(t *testing.T, hub *hub) *session {
	t.Helper()
	engine := newRtpEngineMock()
	s := newSession(uuid.New(), hub, engine, onQuitSessionInternallyStub)
	hub.sessionRepo.Add(s)
	return s
}

func TestHub(t *testing.T) {

	t.Run("get tracks from other session", func(t *testing.T) {
		hub := testHubSetup(t)
		defer hub.stop()
		assert.NotNil(t, hub)
	})
}
