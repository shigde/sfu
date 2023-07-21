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
	lobby := newRtpStreamLobby(uuid.New(), engine)
	return lobby
}
func TestStreamLobby(t *testing.T) {

	t.Run("join lobby", func(t *testing.T) {
		lobby := testStreamLobbySetup(t)
		defer lobby.stop()

		joinRequest := newJoinRequest(context.Background(), uuid.New(), mockedOffer)

		go lobby.runJoin(joinRequest)

		select {
		case data := <-joinRequest.response:
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
		joinRequest := newJoinRequest(ctx, uuid.New(), mockedOffer)

		cancel()
		go lobby.runJoin(joinRequest)

		select {
		case err := <-joinRequest.err:
			assert.ErrorIs(t, err, errLobbyRequestTimeout)
		case <-time.After(time.Second * 3):
			t.Fail()
		}
	})

}
