package sessions

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/mocks"
	"github.com/stretchr/testify/assert"
)

func testHubSetup(t *testing.T) (*Hub, func()) {
	t.Helper()
	sessions := NewSessionRepository()
	engine := mocks.NewRtpEngine()
	forwarder := mocks.NewLiveSender()
	ctx, cancel := context.WithCancel(context.Background())
	hub := NewHub(ctx, sessions, uuid.New(), forwarder)
	s1 := NewSession(ctx, uuid.New(), hub, engine, UserSession, nil)
	s2 := NewSession(ctx, uuid.New(), hub, engine, UserSession, nil)
	sessions.Add(s1)
	sessions.Add(s2)
	return hub, cancel
}
func testHubSessionSetup(t *testing.T, hub *Hub) *Session {
	t.Helper()
	engine := mocks.NewRtpEngine()
	s := NewSession(hub.ctx, uuid.New(), hub, engine, UserSession, nil)
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
