package lobby

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

// Todo refactor this to cmd pattern
type lobbyRequest struct {
	user uuid.UUID
	ctx  context.Context
	err  chan error
	data interface{}
}

type createIngressEndpointData struct {
	offer    *webrtc.SessionDescription
	response chan *createIngressEndpointResponse
}

type initEgressEndpointData struct {
	response chan *initEgressEndpointResponse
}

type finalCreateEgressEndpointData struct {
	answer   *webrtc.SessionDescription
	response chan *finalCreateEgressEndpointResponse
}

type leaveData struct {
	response chan bool
}

type liveStreamData struct {
	cmd      string
	key      string
	rtmpUrl  string
	response chan bool
}

func newLobbyRequest(ctx context.Context, user uuid.UUID) *lobbyRequest {
	errChan := make(chan error)
	return &lobbyRequest{
		user: user,
		err:  errChan,
		ctx:  ctx,
	}
}

func newIngressEndpointData(offer *webrtc.SessionDescription) *createIngressEndpointData {
	resChan := make(chan *createIngressEndpointResponse)
	return &createIngressEndpointData{
		offer:    offer,
		response: resChan,
	}
}

func newInitEgressEndpointData() *initEgressEndpointData {
	resChan := make(chan *initEgressEndpointResponse)
	return &initEgressEndpointData{
		response: resChan,
	}
}

func newFinalCreateEgressEndpointData(answer *webrtc.SessionDescription) *finalCreateEgressEndpointData {
	resChan := make(chan *finalCreateEgressEndpointResponse)
	return &finalCreateEgressEndpointData{
		answer:   answer,
		response: resChan,
	}
}

func newLiveStreamStart(key string, rtmpUrl string) *liveStreamData {
	resChan := make(chan bool)
	return &liveStreamData{
		cmd:      "start",
		key:      key,
		rtmpUrl:  rtmpUrl,
		response: resChan,
	}
}

func newLiveStreamStop() *liveStreamData {
	resChan := make(chan bool)
	return &liveStreamData{
		cmd:      "stop",
		response: resChan,
	}
}

func newLeaveData() *leaveData {
	resChan := make(chan bool)
	return &leaveData{
		response: resChan,
	}
}

type createIngressEndpointResponse struct {
	answer       *webrtc.SessionDescription
	resource     uuid.UUID
	RtpSessionId uuid.UUID
}

type initEgressEndpointResponse struct {
	offer        *webrtc.SessionDescription
	RtpSessionId uuid.UUID
}

type finalCreateEgressEndpointResponse struct {
	RtpSessionId uuid.UUID
}
