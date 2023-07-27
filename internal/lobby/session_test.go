package lobby

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/logging"
	"github.com/stretchr/testify/assert"
)

func testRtpSessionSetup(t *testing.T) (*session, *rtpEngineMock) {
	t.Helper()
	logging.SetupDebugLogger()
	engine := mockRtpEngineForOffer(mockedAnswer)

	session := newSession(uuid.New(), &hub{}, engine)
	return session, engine
}
func TestRtpSessionOffer(t *testing.T) {
	t.Run("offer a sessions after sessions was stopped", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		offerReq := newOfferRequest(ctx, offer, offerTypeReceving)
		_ = session.stop()
		go session.runOfferRequest(offerReq)

		select {
		case <-offerReq.answer:
			t.Fatalf("No answer was expected!")
		case <-offerReq.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-offerReq.err:
			assert.ErrorIs(t, err, errRtpSessionAlreadyClosed)
		}
	})

	t.Run("offer a sessions and receive an answer", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		offerReq := newOfferRequest(context.Background(), mockedOffer, offerTypeReceving)
		go func() {
			session.runOfferRequest(offerReq)
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
	t.Run("stop sessions right after start sessions", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		err := session.stop()
		assert.NoError(t, err)
	})

	t.Run("stop sessions twice time not possible", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		err := session.stop()
		assert.NoError(t, err)
		err = session.stop()
		assert.ErrorIs(t, err, errRtpSessionAlreadyClosed)
	})
}
