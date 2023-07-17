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
	engine := newRtpEngineMock()

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
		offer := &webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: "--o--"}
		answer := &webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "--a--"}
		session, engine := testRtpSessionSetup(t)
		engine.conn = mockConnection(answer)

		offerReq := newOfferRequest(context.Background(), offer)
		go func() {
			session.runOffer(offerReq)
		}()
		select {
		case res := <-offerReq.answer:
			assert.Equal(t, res, answer)
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
