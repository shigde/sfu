package lobby

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/logging"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/stretchr/testify/assert"
)

func testRtpSessionSetup(t *testing.T) (*session, *rtpEngineMock) {
	t.Helper()
	logging.SetupDebugLogger()
	engine := mockRtpEngineForOffer(mockedAnswer)
	forwarder := newStreamForwarderMock()
	hub := newHub(newSessionRepository(), forwarder, nil)
	session := newSession(uuid.New(), hub, engine, nil)
	return session, engine
}

func TestRtpSessionOffer(t *testing.T) {
	t.Run("offerIngressReq to session after session was stopped", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := newSessionRequest(ctx, offer, offerIngressReq)
		_ = session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerEgressReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, errSessionAlreadyClosed)
		}
	})

	t.Run("offerIngressReq to session but receiver already exists", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)

		session.receiver = newReceiverHandler(session.Id, uuid.New(), nil)
		session.receiver.endpoint = mockConnection(nil)

		req := newSessionRequest(context.Background(), offer, offerIngressReq)
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerEgressReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, errReceiverInSessionAlreadyExists)
		}
	})

	t.Run("offerIngressReq to session and receive an answer", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := newSessionRequest(context.Background(), mockedOffer, offerIngressReq)
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

	t.Run("offerIngressReq to session but receive an ice gathering timeout", func(t *testing.T) {
		session, engine := testRtpSessionSetup(t)
		engine.conn = mockIdelConnection()

		req := newSessionRequest(context.Background(), mockedOffer, offerIngressReq)
		before := iceGatheringTimeout
		iceGatheringTimeout = 0
		go func() {
			session.runRequest(req)
		}()
		select {
		case _ = <-req.respSDPChan:
			t.Fatalf("No answer was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, rtp.ErrIceGatheringInterruption)
		}
		iceGatheringTimeout = before
	})
}

func TestRtpSessionStartListen(t *testing.T) {
	t.Run("initEgressReq to session after sessions was stopped", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := newSessionRequest(ctx, nil, initEgressReq)
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

	t.Run("initEgressReq to session but sender already exists", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		session.sender = newSenderHandler(session.Id, uuid.New(), newMockedMessenger(t))
		session.sender.endpoint = mockConnection(nil)

		req := newSessionRequest(context.Background(), nil, initEgressReq)
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

	t.Run("initEgressReq to session and receive an offer but no receiver exists", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := newSessionRequest(context.Background(), nil, initEgressReq)
		go func() {
			session.runRequest(req)
		}()

		select {
		case <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, errNoReceiverInSession)
		}
	})

	t.Run("initEgressReq to session and receive an offer but the receiver has no messenger", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		session.receiver = newReceiverHandler(session.Id, session.user, nil)
		req := newSessionRequest(context.Background(), nil, initEgressReq)

		oldTimeOut := waitingTimeOut
		waitingTimeOut = 0
		go func() {
			session.runRequest(req)
		}()
		select {
		case <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, errTimeoutByWaitingForMessenger)
		}
		waitingTimeOut = oldTimeOut
	})

	t.Run("initEgressReq to session and receive an offer", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := newSessionRequest(context.Background(), nil, initEgressReq)
		session.receiver = newReceiverHandler(session.Id, session.user, nil)
		session.receiver.messenger = newMockedMessenger(t)
		session.receiver.stopWaitingForMessenger()

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

	t.Run("initEgressReq to session but receive an ice gathering timeout", func(t *testing.T) {
		session, engine := testRtpSessionSetup(t)
		engine.conn = mockIdelConnection()

		req := newSessionRequest(context.Background(), nil, initEgressReq)
		session.receiver = newReceiverHandler(session.Id, session.user, nil)
		session.receiver.messenger = newMockedMessenger(t)
		session.receiver.stopWaitingForMessenger()

		before := iceGatheringTimeout
		iceGatheringTimeout = 0
		go func() {
			session.runRequest(req)
		}()
		select {
		case _ = <-req.respSDPChan:
			t.Fatalf("No answer was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, rtp.ErrIceGatheringInterruption)
		}
		iceGatheringTimeout = before
	})
}

func TestRtpSessionListen(t *testing.T) {
	t.Run("answerEgressReq to session after sessions was stopped", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := newSessionRequest(ctx, mockedAnswer, answerEgressReq)
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

	t.Run("answerEgressReq to session without a sending connection", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := newSessionRequest(context.Background(), mockedAnswer, answerEgressReq)
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

	t.Run("answerEgressReq to session", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		session.sender = newSenderHandler(session.Id, uuid.New(), newMockedMessenger(t))
		session.sender.endpoint = mockConnection(mockedOffer)

		req := newSessionRequest(context.Background(), mockedAnswer, answerEgressReq)
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
