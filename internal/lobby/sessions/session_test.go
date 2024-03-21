package sessions

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby"
	"github.com/shigde/sfu/internal/lobby/clients"
	"github.com/shigde/sfu/internal/lobby/mocks"
	"github.com/shigde/sfu/internal/logging"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/stretchr/testify/assert"
)

func testRtpSessionSetup(t *testing.T) (*Session, *mocks.RtpEngineMock) {
	t.Helper()
	logging.SetupDebugLogger()
	engine := mocks.NewRtpEngineForOffer(mocks.Answer)
	forwarder := mocks.NewLiveSender()
	ctx, _ := context.WithCancel(context.Background())
	hub := NewHub(ctx, NewSessionRepository(), uuid.New(), forwarder)
	session := NewSession(ctx, uuid.New(), hub, engine, nil)
	return session, engine
}

func TestRtpSessionOffer(t *testing.T) {
	t.Run("offerIngressReq to session after session was stopped", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := lobby.newSessionRequest(ctx, offer, lobby.offerIngressReq)
		session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerEgressReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errSessionAlreadyClosed)
		}
	})

	t.Run("offerIngressReq to session but receiver already exists", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		session.ingress = lobby.mockConnection(nil)

		req := lobby.newSessionRequest(context.Background(), offer, lobby.offerIngressReq)
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerEgressReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errReceiverInSessionAlreadyExists)
		}
	})

	t.Run("offerIngressReq to session and receive an answer", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := lobby.newSessionRequest(context.Background(), lobby.mockedOffer, lobby.offerIngressReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case res := <-req.respSDPChan:
			assert.Equal(t, res, lobby.mockedAnswer)
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case <-req.err:
			t.Fatalf("No error was expected!")
		}
	})

	t.Run("offerIngressReq to session but receive an ice gathering timeout", func(t *testing.T) {
		session, engine := testRtpSessionSetup(t)
		engine.conn = lobby.mockIdelConnection()

		req := lobby.newSessionRequest(context.Background(), lobby.mockedOffer, lobby.offerIngressReq)
		before := lobby.iceGatheringTimeout
		lobby.iceGatheringTimeout = 0
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
		lobby.iceGatheringTimeout = before
	})
}

func TestRtpSessionStartListen(t *testing.T) {
	t.Run("initEgressReq to session after sessions was stopped", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := lobby.newSessionRequest(ctx, nil, lobby.initEgressReq)
		session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errSessionAlreadyClosed)
		}
	})

	t.Run("initEgressReq to session but sender already exists", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		session.egress = lobby.mockConnection(nil)

		req := lobby.newSessionRequest(context.Background(), nil, lobby.initEgressReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case _ = <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errSenderInSessionAlreadyExists)
		}
	})

	t.Run("initEgressReq to session and receive an offer but no receiver exists", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := lobby.newSessionRequest(context.Background(), nil, lobby.initEgressReq)
		go func() {
			session.runRequest(req)
		}()

		select {
		case <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errNoReceiverInSession)
		}
	})

	t.Run("initEgressReq to session and receive an offer but the signal has no messenger", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		session.ingress = lobby.mockConnection(nil)
		req := lobby.newSessionRequest(context.Background(), nil, lobby.initEgressReq)

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
		req := lobby.newSessionRequest(context.Background(), nil, lobby.initEgressReq)
		session.ingress = lobby.mockConnection(nil)
		session.signal.messenger = clients.newMockedMessenger(t)
		session.signal.stopWaitingForMessenger()

		go func() {
			session.runRequest(req)
		}()
		select {
		case res := <-req.respSDPChan:
			assert.Equal(t, res, lobby.mockedAnswer)
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case <-req.err:
			t.Fatalf("No error was expected!")
		}
	})

	t.Run("initEgressReq to session but receive an ice gathering timeout", func(t *testing.T) {
		session, engine := testRtpSessionSetup(t)
		engine.conn = lobby.mockIdelConnection()

		req := lobby.newSessionRequest(context.Background(), nil, lobby.initEgressReq)
		session.ingress = lobby.mockConnection(nil)
		session.signal.messenger = clients.newMockedMessenger(t)
		session.signal.stopWaitingForMessenger()

		before := lobby.iceGatheringTimeout
		lobby.iceGatheringTimeout = 0
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
		lobby.iceGatheringTimeout = before
	})
}

func TestRtpSessionListen(t *testing.T) {
	t.Run("answerEgressReq to session after sessions was stopped", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := lobby.newSessionRequest(ctx, lobby.mockedAnswer, lobby.answerEgressReq)
		session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errSessionAlreadyClosed)
		}
	})

	t.Run("answerEgressReq to session without a sending connection", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := lobby.newSessionRequest(context.Background(), lobby.mockedAnswer, lobby.answerEgressReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case _ = <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errNoSenderInSession)
		}
	})

	t.Run("answerEgressReq to session", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		session.ingress = lobby.mockConnection(nil)
		session.signal.messenger = clients.newMockedMessenger(t)
		session.egress = lobby.mockConnection(lobby.mockedOffer)
		session.signal.egress = session.egress

		req := lobby.newSessionRequest(context.Background(), lobby.mockedAnswer, lobby.answerEgressReq)
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
		session.stop()
		assert.Equal(t, <-session.ctx.Done(), struct{}{})
	})

	//t.Run("stop sessions twice time not possible", func(t *testing.T) {
	//	session, _ := testRtpSessionSetup(t)
	//	err := session.stop()
	//	assert.NoError(t, err)
	//	err = session.stop()
	//	assert.ErrorIs(t, err, errSessionAlreadyClosed)
	//})
}

func TestRtpSessionOfferStaticEndpoint(t *testing.T) {
	t.Run("offerStaticEgressReq to closed session", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := lobby.newSessionRequest(ctx, offer, lobby.offerStaticEgressReq)
		session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerEgressReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errSessionAlreadyClosed)
		}
	})

	t.Run("offerStaticEgressReq session but receive an ice gathering timeout", func(t *testing.T) {
		session, engine := testRtpSessionSetup(t)
		engine.conn = lobby.mockIdelConnection()
		before := lobby.iceGatheringTimeout
		lobby.iceGatheringTimeout = 0

		req := lobby.newSessionRequest(context.Background(), lobby.mockedOffer, lobby.offerStaticEgressReq)
		go session.runRequest(req)

		select {
		case _ = <-req.respSDPChan:
			t.Fatalf("No answer was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, rtp.ErrIceGatheringInterruption)
		}
		lobby.iceGatheringTimeout = before
	})

	t.Run("offerStaticEgressReq session", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)

		req := lobby.newSessionRequest(context.Background(), lobby.mockedOffer, lobby.offerStaticEgressReq)
		go session.runRequest(req)

		select {
		case res := <-req.respSDPChan:
			assert.Equal(t, res, lobby.mockedAnswer)
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case <-req.err:
			t.Fatalf("No error was expected!")
		}
	})
}
func TestRtpSessionHostEgressOffer(t *testing.T) {
	t.Run("hostOfferEgressReq to session after session was stopped", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := lobby.newSessionRequest(ctx, offer, lobby.offerHostEgressReq)
		session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerEgressReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errSessionAlreadyClosed)
		}
	})

	t.Run("hostOfferEgressReq to session but session already exists", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		session.egress = lobby.mockConnection(nil)

		req := lobby.newSessionRequest(context.Background(), offer, lobby.offerHostEgressReq)
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerEgressReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errSenderInSessionAlreadyExists)
		}
	})

	t.Run("offerHostEgressReq to session and receive an offer", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := lobby.newSessionRequest(context.Background(), nil, lobby.offerHostEgressReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case res := <-req.respSDPChan:
			assert.Equal(t, res, lobby.mockedAnswer)
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case <-req.err:
			t.Fatalf("No error was expected!")
		}
	})

	t.Run("offerHostEgressReq to session but receive an ice gathering timeout", func(t *testing.T) {
		session, engine := testRtpSessionSetup(t)
		engine.conn = lobby.mockIdelConnection()

		req := lobby.newSessionRequest(context.Background(), nil, lobby.offerHostEgressReq)
		before := lobby.iceGatheringTimeout
		lobby.iceGatheringTimeout = 0
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
		lobby.iceGatheringTimeout = before
	})
}

func TestRtpSessionHostEgressAnswer(t *testing.T) {
	t.Run("answerHostEgressReq to session after sessions was stopped", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := lobby.newSessionRequest(ctx, lobby.mockedAnswer, lobby.answerHostEgressReq)
		session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errSessionAlreadyClosed)
		}
	})

	t.Run("answerHostEgressReq to session without a sending connection", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := lobby.newSessionRequest(context.Background(), lobby.mockedAnswer, lobby.answerHostEgressReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case _ = <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errNoSenderInSession)
		}
	})

	t.Run("answerHostEgressReq without messenger", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		session.signal.messenger = clients.newMockedMessenger(t)
		session.egress = lobby.mockConnection(lobby.mockedOffer)
		session.signal.egress = session.egress

		req := lobby.newSessionRequest(context.Background(), lobby.mockedAnswer, lobby.answerHostEgressReq)
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

	t.Run("answerHostEgressReq to session", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		session.signal.messenger = clients.newMockedMessenger(t)
		session.signal.stopWaitingForMessenger()
		session.egress = lobby.mockConnection(lobby.mockedOffer)
		session.signal.egress = session.egress

		req := lobby.newSessionRequest(context.Background(), lobby.mockedAnswer, lobby.answerHostEgressReq)
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

func TestRtpSessionHostIngress(t *testing.T) {
	t.Run("offerHostIngressReq to session after session was stopped", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := lobby.newSessionRequest(ctx, offer, lobby.offerHostIngressReq)
		session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerEgressReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errSessionAlreadyClosed)
		}
	})

	t.Run("offerHostIngressReq to session but receiver already exists", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		session.ingress = lobby.mockConnection(nil)

		req := lobby.newSessionRequest(context.Background(), offer, lobby.offerHostIngressReq)
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerEgressReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errReceiverInSessionAlreadyExists)
		}
	})

	t.Run("offerHostIngressReq to session and receive an answer", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := lobby.newSessionRequest(context.Background(), lobby.mockedOffer, lobby.offerHostIngressReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case res := <-req.respSDPChan:
			assert.Equal(t, res, lobby.mockedAnswer)
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case <-req.err:
			t.Fatalf("No error was expected!")
		}
	})

	t.Run("offerHostIngressReq to session but receive an ice gathering timeout", func(t *testing.T) {
		session, engine := testRtpSessionSetup(t)
		engine.conn = lobby.mockIdelConnection()

		req := lobby.newSessionRequest(context.Background(), lobby.mockedOffer, lobby.offerHostIngressReq)
		before := lobby.iceGatheringTimeout
		lobby.iceGatheringTimeout = 0
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
		lobby.iceGatheringTimeout = before
	})
}

func TestRtpSessionHostPipeOffer(t *testing.T) {
	t.Run("offerHostPipeReq to session after session was stopped", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := lobby.newSessionRequest(ctx, offer, lobby.offerHostPipeReq)
		session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerEgressReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errSessionAlreadyClosed)
		}
	})

	t.Run("offerHostPipeReq to session but session already exists", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		session.channel = lobby.mockConnection(nil)

		req := lobby.newSessionRequest(context.Background(), offer, lobby.offerHostPipeReq)
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerEgressReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errSenderInSessionAlreadyExists)
		}
	})

	t.Run("offerHostPipeReq to session and receive an offer", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := lobby.newSessionRequest(context.Background(), nil, lobby.offerHostPipeReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case res := <-req.respSDPChan:
			assert.Equal(t, res, lobby.mockedAnswer)
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case <-req.err:
			t.Fatalf("No error was expected!")
		}
	})

	t.Run("offerHostPipeReq to session but receive an ice gathering timeout", func(t *testing.T) {
		session, engine := testRtpSessionSetup(t)
		engine.conn = lobby.mockIdelConnection()

		req := lobby.newSessionRequest(context.Background(), nil, lobby.offerHostPipeReq)
		before := lobby.iceGatheringTimeout
		lobby.iceGatheringTimeout = 0
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
		lobby.iceGatheringTimeout = before
	})
}

func TestRtpSessionHostPipeAnswer(t *testing.T) {
	t.Run("answerHostPipeReq to session after sessions was stopped", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := lobby.newSessionRequest(ctx, lobby.mockedAnswer, lobby.answerHostPipeReq)
		session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errSessionAlreadyClosed)
		}
	})

	t.Run("answerHostPipeReq to session without a sending connection", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := lobby.newSessionRequest(context.Background(), lobby.mockedAnswer, lobby.answerHostPipeReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case _ = <-req.respSDPChan:
			t.Fatalf("No sdp was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errNoSenderInSession)
		}
	})

	t.Run("answerHostPipeReq to session", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		session.signal.messenger = clients.newMockedMessenger(t)
		session.channel = lobby.mockConnection(lobby.mockedOffer)

		req := lobby.newSessionRequest(context.Background(), lobby.mockedAnswer, lobby.answerHostPipeReq)
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

func TestRtpSessionHostRemotePipe(t *testing.T) {
	t.Run("offerHostRemotePipeReq to session after session was stopped", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		ctx := context.Background()
		req := lobby.newSessionRequest(ctx, offer, lobby.offerHostRemotePipeReq)
		session.stop()
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerPipeReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errSessionAlreadyClosed)
		}
	})

	t.Run("offerHostRemotePipeReq to session but channel already exists", func(t *testing.T) {
		var offer *webrtc.SessionDescription
		session, _ := testRtpSessionSetup(t)
		session.channel = lobby.mockConnection(nil)

		req := lobby.newSessionRequest(context.Background(), offer, lobby.offerHostRemotePipeReq)
		go session.runRequest(req)

		select {
		case <-req.respSDPChan:
			t.Fatalf("No answerPipeReq was expected!")
		case <-req.ctx.Done():
			t.Fatalf("No canceling was expected!")
		case err := <-req.err:
			assert.ErrorIs(t, err, lobby.errReceiverInSessionAlreadyExists)
		}
	})

	t.Run("offerHostRemotePipeReq to session and receive an answer", func(t *testing.T) {
		session, _ := testRtpSessionSetup(t)
		req := lobby.newSessionRequest(context.Background(), lobby.mockedOffer, lobby.offerHostRemotePipeReq)
		go func() {
			session.runRequest(req)
		}()
		select {
		case res := <-req.respSDPChan:
			assert.Equal(t, res, lobby.mockedAnswer)
		case <-req.ctx.Done():
			t.Fatalf("No cancel was expected!")
		case <-req.err:
			t.Fatalf("No error was expected!")
		}
	})

	t.Run("offerHostRemotePipeReq to session but receive an ice gathering timeout", func(t *testing.T) {
		session, engine := testRtpSessionSetup(t)
		engine.conn = lobby.mockIdelConnection()

		req := lobby.newSessionRequest(context.Background(), lobby.mockedOffer, lobby.offerHostRemotePipeReq)
		before := lobby.iceGatheringTimeout
		lobby.iceGatheringTimeout = 0
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
		lobby.iceGatheringTimeout = before
	})
}
