package lobby

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/logging"
	"github.com/stretchr/testify/assert"
)

func testRtpSessionSetup(t *testing.T) *rtpSession {
	t.Helper()
	logging.SetupDebugLogger()
	engine := newRtpEngineMock()
	session := newRtpSession(uuid.New(), engine)
	return session
}
func TestRtpSessionOffer(t *testing.T) {
	t.Run("offer a session after session was stopped", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session := testRtpSessionSetup(t)
		ctx := context.Background()
		offerReq := newOfferRequest(ctx, offer)
		_ = session.stop()
		go session.runOffer(offerReq)

		select {
		case <-offerReq.answer:
			t.Fatalf("No answer was expected!")
		case <-offerReq.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-offerReq.err:
			assert.ErrorIs(t, err, errRtpSessionAlreadyClosed)
		}
	})

	t.Run("offer a session and cancel request", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session := testRtpSessionSetup(t)
		ctx, cancel := context.WithCancel(context.Background())
		offerReq := newOfferRequest(ctx, offer)
		cancel()
		go func() {
			session.runOffer(offerReq)
		}()
		select {
		case <-offerReq.answer:
			t.Fatalf("No answer was expected!")
		case <-offerReq.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-offerReq.err:
			assert.ErrorIs(t, err, errRtpSessionAlreadyClosed)
		}
	})
}

func TestRtpSessionStop(t *testing.T) {
	t.Run("stop session right after start session", func(t *testing.T) {
		session := testRtpSessionSetup(t)
		err := session.stop()
		assert.NoError(t, err)
	})

	t.Run("stop session twice time not possible", func(t *testing.T) {
		session := testRtpSessionSetup(t)
		err := session.stop()
		assert.NoError(t, err)
		err = session.stop()
		assert.ErrorIs(t, err, errRtpSessionAlreadyClosed)
	})
}
