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

	session := newSession(uuid.New(), newHub(newSessionRepository()), engine, onQuitSessionInternallyStub)
	return session, engine
}

func TestRtpSessionOffer(t *testing.T) {
	t.Run("offerReq to session after session was stopped", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := newSessionRequest(ctx, offer, offerReq)
		_ = session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, errSessionAlreadyClosed)
		}
	})

	t.Run("offerReq to session but receiver already exists", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)

		session.receiver = newReceiverHandler(session.Id, uuid.New(), nil)
		session.receiver.endpoint = mockConnection(nil)

		req := newSessionRequest(context.Background(), offer, offerReq)
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, errReceiverInSessionAlreadyExists)
		}
	})

	t.Run("offerReq to session and receive an answer", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := newSessionRequest(context.Background(), mockedOffer, offerReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case res := <-req.respSDPChan:
			assert.Equal(t, res, mockedAnswer)
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case <-req.err:
			t.Fatalf("No error was expected!")
		}
	})
}

func TestRtpSessionStartListen(t *testing.T) {
	t.Run("startReq to session after sessions was stopped", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := newSessionRequest(ctx, nil, startReq)
		_ = session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, errSessionAlreadyClosed)
		}
	})

	t.Run("startReq to session but sender already exists", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		session.sender = newSenderHandler(session.Id, uuid.New(), nil)
		session.sender.endpoint = mockConnection(nil)

		req := newSessionRequest(context.Background(), nil, startReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case _ = <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, errSenderInSessionAlreadyExists)
		}
	})

	t.Run("startReq to session and receive an offer but no receiver exists", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := newSessionRequest(context.Background(), nil, startReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, errReceiverInSessionHasNoMessenger)
		}
	})

	t.Run("startReq to session and receive an offer but the receiver has no messenger", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		session.receiver = newReceiverHandler(session.Id, session.user, nil)
		req := newSessionRequest(context.Background(), nil, startReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, errReceiverInSessionHasNoMessenger)
		}
	})

	t.Run("startReq to session and receive an offer", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := newSessionRequest(context.Background(), nil, startReq)
		session.receiver = newReceiverHandler(session.Id, session.user, nil)
		session.receiver.messenger = newMessenger(nil)
		go func() {
			session.runRequest(req)
		}()
		select {
		case res := <-req.respSDPChan:
			assert.Equal(t, res, mockedAnswer)
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case <-req.err:
			t.Fatalf("No error was expected!")
		}
	})
}

func TestRtpSessionListen(t *testing.T) {
	t.Run("answerReq to session after sessions was stopped", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := newSessionRequest(ctx, mockedAnswer, answerReq)
		_ = session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, errSessionAlreadyClosed)
		}
	})

	t.Run("answerReq to session without a sending connection", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := newSessionRequest(context.Background(), mockedAnswer, answerReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case _ = <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, errNoSenderInSession)
		}
	})

	t.Run("answerReq to session", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		session.sender = newSenderHandler(session.Id, uuid.New(), nil)
		session.sender.endpoint = mockConnection(mockedOffer)

		req := newSessionRequest(context.Background(), mockedAnswer, answerReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case res := <-req.respSDPChan:
			assert.Nil(t, res)
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case <-req.err:
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

	//t.Run("stop sessions twice time not possible", func(t *testing.T) {
	//	session, _ := testRtpSessionSetup(t)
	//	err := session.stop()
	//	assert.NoError(t, err)
	//	err = session.stop()
	//	assert.ErrorIs(t, err, errSessionAlreadyClosed)
	//})
}
