package lobby

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func testHubSetup(t *testing.T) (*hub, func()) {
	t.Helper()
	sessions := newSessionRepository()
	engine := newRtpEngineMock()
	forwarder := newLiveStreamSenderMock()
	ctx, cancel := context.WithCancel(context.Background())
	hub := newHub(ctx, sessions, uuid.New(), forwarder)
	s1 := newSession(uuid.New(), hub, engine, nil)
	s2 := newSession(uuid.New(), hub, engine, nil)
	sessions.Add(s1)
	sessions.Add(s2)
	return hub, cancel
}
func testHubSessionSetup(t *testing.T, hub *hub) *session {
	t.Helper()
	engine := newRtpEngineMock()
	s := newSession(uuid.New(), hub, engine, nil)
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
