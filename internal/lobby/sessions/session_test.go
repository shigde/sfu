package sessions

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/mocks"
	"github.com/shigde/sfu/internal/logging"
	"github.com/stretchr/testify/assert"
)

var (
	runtimeProcessWaitingTimeout = processWaitingTimeout
)

func testSessionSetup(t *testing.T) (*Session, *mocks.RtpEngineMock) {
	t.Helper()
	logging.SetupDebugLogger()
	engine := mocks.NewRtpEngineForOffer(mocks.Answer)
	forwarder := mocks.NewLiveSender()
	ctx, _ := context.WithCancel(context.Background())
	hub := NewHub(ctx, NewSessionRepository(), uuid.New(), forwarder)
	session := NewSession(ctx, uuid.New(), hub, engine, nil)
	return session, engine
}

func testSetupEgress(t *testing.T, session *Session) {
	t.Helper()
	close(session.signal.receivedMessenger)
	// @TODO: The signaling should be independent of ingress and egress
	session.ingress = mocks.NewEndpoint(nil)
	session.signal.ingress = session.ingress
}

func TestSession_CreateIngressEndpoint(t *testing.T) {
	t.Run("after session was stopped", func(t *testing.T) {
		session, _ := testSessionSetup(t)
		session.stop()

		_, err := session.CreateIngressEndpoint(context.Background(), mocks.Offer)
		assert.ErrorIs(t, err, ErrSessionAlreadyClosed)
	})

	t.Run("if endpoint already exists", func(t *testing.T) {
		session, _ := testSessionSetup(t)
		session.ingress = mocks.NewEndpoint(nil)

		_, err := session.CreateIngressEndpoint(context.Background(), mocks.Offer)
		assert.ErrorIs(t, err, ErrIngressAlreadyExists)
	})

	t.Run("if endpoint fails", func(t *testing.T) {
		session, engine := testSessionSetup(t)
		engine.Err = errors.New("fail")

		_, err := session.CreateIngressEndpoint(context.Background(), mocks.Offer)
		assert.ErrorContains(t, err, "create rtp endpoint")
	})

	t.Run("if answer fails", func(t *testing.T) {
		session, engine := testSessionSetup(t)
		engine.Conn = mocks.NewIdelEndpoint() // Ice gathering not complete
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Force to Run in time out because waiting for ice gathering complete

		_, err := session.CreateIngressEndpoint(ctx, mocks.Offer)
		assert.ErrorContains(t, err, "create ingress answer resource")
	})

	t.Run("get answer", func(t *testing.T) {
		session, _ := testSessionSetup(t)

		answer, _ := session.CreateIngressEndpoint(context.Background(), mocks.Offer)
		assert.Equal(t, mocks.Answer, answer)
	})
}

func TestSession_CreateEgressEndpoint(t *testing.T) {
	t.Run("after session was stopped", func(t *testing.T) {
		session, _ := testSessionSetup(t)
		session.stop()

		_, err := session.CreateEgressEndpoint(context.Background(), mocks.Offer)
		assert.ErrorIs(t, err, ErrSessionAlreadyClosed)
	})

	t.Run("if endpoint already exists", func(t *testing.T) {
		session, _ := testSessionSetup(t)
		session.ingress = mocks.NewEndpoint(nil)
		session.egress = mocks.NewEndpoint(nil)

		_, err := session.CreateEgressEndpoint(context.Background(), mocks.Offer)
		assert.ErrorIs(t, err, ErrEgressAlreadyExists)
	})

	t.Run("if signal channel endpoint not exists", func(t *testing.T) {
		session, _ := testSessionSetup(t)
		_, err := session.CreateEgressEndpoint(context.Background(), mocks.Offer)
		assert.ErrorIs(t, err, ErrNoSignalChannel)
	})

	t.Run("if no signal messanger", func(t *testing.T) {
		session, engine := testSessionSetup(t)
		session.ingress = mocks.NewEndpoint(nil)
		processWaitingTimeout = 0 // Force to Run in time out because signaling setup is not complete
		engine.Err = errors.New("fail")

		_, err := session.CreateEgressEndpoint(context.Background(), mocks.Offer)
		assert.ErrorIs(t, err, ErrSessionProcessWaitingTimeout)
		processWaitingTimeout = runtimeProcessWaitingTimeout
	})

	t.Run("if endpoint fails", func(t *testing.T) {
		session, engine := testSessionSetup(t)
		testSetupEgress(t, session)
		engine.Err = errors.New("fail")

		_, err := session.CreateEgressEndpoint(context.Background(), mocks.Offer)
		assert.ErrorContains(t, err, "create rtp endpoint")
	})

	t.Run("if answer fails", func(t *testing.T) {
		session, engine := testSessionSetup(t)
		testSetupEgress(t, session)
		engine.Conn = mocks.NewIdelEndpoint() // Ice gathering not complete
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Force to Run in timeout because waiting for ice gathering complete

		_, err := session.CreateEgressEndpoint(ctx, mocks.Offer)
		assert.ErrorContains(t, err, "create egress answer resource")
	})

	t.Run("get answer", func(t *testing.T) {
		session, _ := testSessionSetup(t)
		testSetupEgress(t, session)

		answer, _ := session.CreateEgressEndpoint(context.Background(), mocks.Offer)
		assert.Equal(t, mocks.Answer, answer)
	})
}
