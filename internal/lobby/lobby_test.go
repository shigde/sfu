package lobby

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/logging"
	"github.com/stretchr/testify/assert"
)

func testStreamLobbySetup(t *testing.T) *lobby {
	t.Helper()
	logging.SetupDebugLogger()
	engine := mockRtpEngineForOffer(mockedAnswer)
	lobby := newLobby(uuid.New(), engine)
	return lobby
}
func TestStreamLobby(t *testing.T) {

	t.Run("join lobby", func(t *testing.T) {
		lobby := testStreamLobbySetup(t)
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
		lobby := testStreamLobbySetup(t)
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
		lobby := testStreamLobbySetup(t)
		defer lobby.stop()

		request := newLobbyRequest(context.Background(), uuid.New())
		listenData := newListenData(mockedOffer)
		request.data = listenData

		go lobby.runRequest(request)

		select {
		case data := <-listenData.response:
			assert.Equal(t, mockedAnswer, data.answer)
			assert.False(t, uuid.Nil == data.RtpSessionId)
		case <-time.After(time.Second * 3):
			t.Fail()
		}
	})

	t.Run("cancel listen lobby", func(t *testing.T) {
		lobby := testStreamLobbySetup(t)
		defer lobby.stop()

		ctx, cancel := context.WithCancel(context.Background())
		request := newLobbyRequest(ctx, uuid.New())
		listenData := newListenData(mockedOffer)
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
}
