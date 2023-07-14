package lobby

import (
	"testing"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
)

func testRtpSessionSetup(t *testing.T) *rtpSession {
	t.Helper()
	var engine rtpEngine
	session := newRtpSession(uuid.New(), engine)
	return session
}
func TestRtpSession(t *testing.T) {
	t.Run("offer session", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session := testRtpSessionSetup(t)
		_, err := session.offer(offer)

		assert.NoError(t, err)
	})
}
