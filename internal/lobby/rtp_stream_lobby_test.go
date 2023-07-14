package lobby

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
)

func testStreamLobbySetup(t *testing.T) *rtpStreamLobby {
	t.Helper()
	var engine rtpEngine
	lobby := newRtpStreamLobby(uuid.New(), engine)
	return lobby
}
func TestStreamLobby(t *testing.T) {

	t.Run("join lobby", func(t *testing.T) {
		lobby := testStreamLobbySetup(t)
		defer lobby.stop()

		var offer *webrtc.SessionDescription
		jR := &joinRequest{
			offer:    offer,
			user:     uuid.New(),
			error:    make(chan error),
			response: make(chan *joinResponse),
			ctx:      context.Background(),
		}

		lobby.request <- jR

		select {
		case data := <-jR.response:
			assert.Nil(t, data.answer)
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
		var offer *webrtc.SessionDescription
		jR := &joinRequest{
			offer:    offer,
			user:     uuid.New(),
			error:    make(chan error),
			response: make(chan *joinResponse),
			ctx:      ctx,
		}

		cancel()
		lobby.request <- jR

		select {
		case err := <-jR.error:
			assert.ErrorIs(t, err, errLobbyRequestTimeout)
		case <-time.After(time.Second * 3):
			t.Fail()
		}
	})

}
