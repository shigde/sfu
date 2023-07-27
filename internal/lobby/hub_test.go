package lobby

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testHubSetup(t *testing.T) *hub {
	t.Helper()
	sessions := newSessionRepository()
	hub := newHub(sessions)
	return hub
}
func TestHub(t *testing.T) {
	hub := testHubSetup(t)
	defer hub.stop()
	assert.NotNil(t, hub)
}
