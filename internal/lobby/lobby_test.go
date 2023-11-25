package lobby

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/logging"
	"github.com/stretchr/testify/assert"
)

func testStreamLobbySetup(t *testing.T) (*lobby, uuid.UUID) {
	t.Helper()
	logging.SetupDebugLogger()
	// set one session in lobby
	engine := mockRtpEngineForOffer(mockedAnswer)
	lobby := newLobby(uuid.New(), engine)
	user := uuid.New()
	session := newSession(user, lobby.hub, engine, lobby.childQuitChan)
	session.sender = newSenderHandler(session.Id, user, newMockedMessenger(t))
	session.sender.endpoint = mockConnection(mockedAnswer)
	lobby.sessions.Add(session)
	return lobby, user
}
func TestStreamLobby(t *testing.T) {

	t.Run("join lobby", func(t *testing.T) {
		lobby, _ := testStreamLobbySetup(t)
		defer lobby.stop()

		request := newLobbyRequest(context.Background(), uuid.New())
		joinData := newIngressEndpointData(mockedOffer)
		request.data = joinData

		go lobby.runRequest(request)

		select {
		case data := <-joinData.response:
			assert.Equal(t, mockedAnswer, data.answer)
			assert.False(t, uuid.Nil == data.resource)
			assert.False(t, uuid.Nil == data.RtpSessionId)
		case <-time.After(time.Second * 3):
			t.Fail()
		}
	})

	t.Run("cancel join lobby", func(t *testing.T) {
		lobby, _ := testStreamLobbySetup(t)
		defer lobby.stop()

		ctx, cancel := context.WithCancel(context.Background())
		request := newLobbyRequest(ctx, uuid.New())
		joinData := newIngressEndpointData(mockedOffer)
		request.data = joinData

		cancel()
		go lobby.runRequest(request)

		select {
		case err := <-request.err:
			assert.ErrorIs(t, err, errLobbyRequestTimeout)
		case <-time.After(time.Second * 3):
			t.Fail()
		}
	})

	t.Run("listen lobby", func(t *testing.T) {
		lobby, user := testStreamLobbySetup(t)
		defer lobby.stop()
		request := newLobbyRequest(context.Background(), user)
		listenData := newListenData(mockedOffer)
		request.data = listenData

		go lobby.runRequest(request)

		select {
		case data := <-listenData.response:
			assert.False(t, uuid.Nil == data.RtpSessionId)
		case <-time.After(time.Second * 3):
			t.Fail()
		}
	})

	t.Run("cancel listen lobby", func(t *testing.T) {
		lobby, user := testStreamLobbySetup(t)
		defer lobby.stop()

		ctx, cancel := context.WithCancel(context.Background())
		request := newLobbyRequest(ctx, user)
		listenData := newListenData(mockedAnswer)
		request.data = listenData

		cancel()
		go lobby.runRequest(request)

		select {
		case err := <-request.err:
			assert.ErrorIs(t, err, errLobbyRequestTimeout)
		case <-time.After(time.Second * 3):
			t.Fatalf("test fails because run in timeout")
		}
	})

	t.Run("start listen lobby but no session was started before", func(t *testing.T) {
		lobby, _ := testStreamLobbySetup(t)
		defer lobby.stop()

		request := newLobbyRequest(context.Background(), uuid.New())
		startData := newInitEgressEndpointData()
		request.data = startData

		go lobby.runRequest(request)

		select {
		case err := <-request.err:
			assert.ErrorIs(t, err, errNoSession)
		case _ = <-startData.response:
			t.Fatalf("test fails because no offer expected")
		case <-time.After(time.Second * 3):
			t.Fatalf("test fails because run in timeout")
		}
	})

	t.Run("start listen lobby session", func(t *testing.T) {
		lobby, _ := testStreamLobbySetup(t)
		defer lobby.stop()

		user := uuid.New()
		session := newSession(user, lobby.hub, mockRtpEngineForOffer(mockedAnswer), lobby.childQuitChan)
		session.receiver = newReceiverHandler(session.Id, session.user, nil)
		session.receiver.messenger = newMockedMessenger(t)
		session.receiver.stopWaitingForMessenger()
		lobby.sessions.Add(session)

		request := newLobbyRequest(context.Background(), user)
		startData := newInitEgressEndpointData()
		request.data = startData

		go lobby.runRequest(request)
		offer := mockedAnswer // its mocked and make no different

		select {
		case err := <-request.err:
			assert.ErrorIs(t, err, errNoSession)
		case data := <-startData.response:
			assert.Equal(t, offer, data.offer)
			assert.False(t, uuid.Nil == data.RtpSessionId)
		case <-time.After(time.Second * 3):
			t.Fatalf("test fails because run in timeout")
		}
	})

	t.Run("cancel start listen lobby", func(t *testing.T) {
		lobby, _ := testStreamLobbySetup(t)
		defer lobby.stop()

		user := uuid.New()
		session := newSession(user, lobby.hub, mockRtpEngineForOffer(mockedAnswer), lobby.childQuitChan)
		session.receiver = newReceiverHandler(session.Id, session.user, nil)
		session.receiver.messenger = newMockedMessenger(t)
		lobby.sessions.Add(session)

		ctx, cancel := context.WithCancel(context.Background())
		request := newLobbyRequest(ctx, user)
		startData := newInitEgressEndpointData()
		request.data = startData

		cancel()
		go lobby.runRequest(request)

		select {
		case err := <-request.err:
			assert.ErrorIs(t, err, errLobbyRequestTimeout)
		case <-time.After(time.Second * 3):
			t.Fatalf("test fails because run in timeout")
		}
	})

	t.Run("stop session internally", func(t *testing.T) {
		lobby, user := testStreamLobbySetup(t)
		defer lobby.stop()
		session, _ := lobby.sessions.FindByUserId(user)

		stopped, _ := lobby.deleteSessionByUserId(user)
		assert.True(t, stopped)
		assert.False(t, lobby.sessions.Contains(session.Id))
	})

	t.Run("leave lobby", func(t *testing.T) {
		lobby, user := testStreamLobbySetup(t)
		defer lobby.stop()
		request := newLobbyRequest(context.Background(), user)
		leaveData := newLeaveData()
		request.data = leaveData

		go lobby.runRequest(request)

		select {
		case success := <-leaveData.response:
			assert.True(t, success)
		case <-time.After(time.Second * 3):
			t.Fatalf("test fails because run in timeout")
		}
	})
}
