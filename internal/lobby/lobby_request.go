package lobby

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type lobbyRequest struct {
	user uuid.UUID
	ctx  context.Context
	err  chan error
	data interface{}
}

type joinData struct {
	offer    *webrtc.SessionDescription
	response chan *joinResponse
}

type startListenData struct {
	response chan *startListenResponse
}

type listenData struct {
	answer   *webrtc.SessionDescription
	response chan *listenResponse
}

type leaveData struct {
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

func newJoinData(offer *webrtc.SessionDescription) *joinData {
	resChan := make(chan *joinResponse)
	return &joinData{
		offer:    offer,
		response: resChan,
	}
}

func newStartListenData() *startListenData {
	resChan := make(chan *startListenResponse)
	return &startListenData{
		response: resChan,
	}
}

func newListenData(answer *webrtc.SessionDescription) *listenData {
	resChan := make(chan *listenResponse)
	return &listenData{
		answer:   answer,
		response: resChan,
	}
}

func newLeaveData() *leaveData {
	resChan := make(chan bool)
	return &leaveData{
		response: resChan,
	}
}

type joinResponse struct {
	answer       *webrtc.SessionDescription
	resource     uuid.UUID
	RtpSessionId uuid.UUID
}

type startListenResponse struct {
	offer        *webrtc.SessionDescription
	RtpSessionId uuid.UUID
}

type listenResponse struct {
	RtpSessionId uuid.UUID
}
