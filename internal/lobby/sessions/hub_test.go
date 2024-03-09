package sessions

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby"
	"github.com/stretchr/testify/assert"
)

func testHubSetup(t *testing.T) (*Hub, func()) {
	t.Helper()
	sessions := lobby.newSessionRepository()
	engine := lobby.newRtpEngineMock()
	forwarder := lobby.newLiveStreamSenderMock()
	ctx, cancel := context.WithCancel(context.Background())
	hub := newHub(ctx, sessions, uuid.New(), forwarder)
	s1 := lobby.newSession(uuid.New(), hub, engine, nil)
	s2 := lobby.newSession(uuid.New(), hub, engine, nil)
	sessions.Add(s1)
	sessions.Add(s2)
	return hub, cancel
}
func testHubSessionSetup(t *testing.T, hub *Hub) *lobby.session {
	t.Helper()
	engine := lobby.newRtpEngineMock()
	s := lobby.newSession(uuid.New(), hub, engine, nil)
	hub.sessionRepo.Add(s)
	return s
}

func TestHub(t *testing.T) {

	t.Run("get tracks from other session", func(t *testing.T) {
		hub, stop := testHubSetup(t)
		defer stop()
		assert.NotNil(t, hub)
	})
}
