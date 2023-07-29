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
	session := newSession(user, lobby.hub, engine)
	session.sender = mockConnection(mockedAnswer)
	lobby.sessions.Add(session)
	return lobby, user
}
func TestStreamLobby(t *testing.T) {

	t.Run("join lobby", func(t *testing.T) {
		lobby, _ := testStreamLobbySetup(t)
		defer lobby.stop()

		request := newLobbyRequest(context.Background(), uuid.New())
		joinData := newJoinData(mockedOffer)
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
		joinData := newJoinData(mockedOffer)
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
			t.Fail()
		}
	})

	t.Run("start listen lobby", func(t *testing.T) {
		lobby, _ := testStreamLobbySetup(t)
		defer lobby.stop()

		request := newLobbyRequest(context.Background(), uuid.New())
		startData := newStartListenData()
		request.data = startData

		go lobby.runRequest(request)
		offer := mockedAnswer // its mocked and make no different

		select {
		case data := <-startData.response:
			assert.Equal(t, offer, data.offer)
			assert.False(t, uuid.Nil == data.RtpSessionId)
		case <-time.After(time.Second * 3):
			t.Fail()
		}
	})

	t.Run("cancel start listen lobby", func(t *testing.T) {
		lobby, _ := testStreamLobbySetup(t)
		defer lobby.stop()

		ctx, cancel := context.WithCancel(context.Background())
		request := newLobbyRequest(ctx, uuid.New())
		startData := newStartListenData()
		request.data = startData

		cancel()
		go lobby.runRequest(request)

		select {
		case err := <-request.err:
			assert.ErrorIs(t, err, errLobbyRequestTimeout)
		case <-time.After(time.Second * 3):
			t.Fail()
		}
	})
}
