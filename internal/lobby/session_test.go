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
	t.Run("offerReq a sessions after sessions was stopped", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		offerReq := newSessionRequest(ctx, offer, offerReq)
		_ = session.stop()
		go session.runRequest(offerReq)

		select {
		case <-offerReq.respSDPChan:
			t.Fatalf("No answerReq was expected!")
		case <-offerReq.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-offerReq.err:
			assert.ErrorIs(t, err, errRtpSessionAlreadyClosed)
		}
	})

	t.Run("offerReq a sessions and receive an answerReq", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		offerReq := newSessionRequest(context.Background(), mockedOffer, offerReq)
		go func() {
			session.runRequest(offerReq)
		}()
		select {
		case res := <-offerReq.respSDPChan:
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
