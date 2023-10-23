package lobby

import (
	"github.com/google/uuid"
)

type streamer struct {
	Id               uuid.UUID
	stream           uuid.UUID
	rtpEngine        rtpEngine
	hub              *hub
	sender           *senderHandler
	quit             chan struct{}
	onInternallyQuit chan<- uuid.UUID
}

func newStreamer(stream uuid.UUID, hub *hub, engine rtpEngine, onInternallyQuit chan<- uuid.UUID) *streamer {
	quit := make(chan struct{})
	streamer := &streamer{
		Id:               uuid.New(),
		stream:           stream,
		rtpEngine:        engine,
		hub:              hub,
		quit:             quit,
		onInternallyQuit: onInternallyQuit,
	}

	return streamer
}
