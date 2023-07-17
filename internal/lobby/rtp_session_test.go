package lobby

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/logging"
	"github.com/stretchr/testify/assert"
)

func testRtpSessionSetup(t *testing.T) (*rtpSession, *rtpEngineMock) {
	t.Helper()
	logging.SetupDebugLogger()
	engine := mockRtpEngineForOffer(mockedAnswer)

	session := newRtpSession(uuid.New(), engine)
	return session, engine
}
func TestRtpSessionOffer(t *testing.T) {
	t.Run("offer a session after session was stopped", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
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

	t.Run("offer a session and receive an answer", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		offerReq := newOfferRequest(context.Background(), mockedOffer)
		go func() {
			session.runOffer(offerReq)
		}()
		select {
		case res := <-offerReq.answer:
			assert.Equal(t, res, mockedAnswer)
		case <-offerReq.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case <-offerReq.err:
			t.Fatalf("No error was expected!")
		}
	})
}

func TestRtpSessionStop(t *testing.T) {
	t.Run("stop session right after start session", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		err := session.stop()
		assert.NoError(t, err)
	})

	t.Run("stop session twice time not possible", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		err := session.stop()
		assert.NoError(t, err)
		err = session.stop()
		assert.ErrorIs(t, err, errRtpSessionAlreadyClosed)
	})
}
