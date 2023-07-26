package lobby

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type joinRequest struct {
	user     uuid.UUID
	offer    *webrtc.SessionDescription
	response chan *joinResponse
	err      chan error
	ctx      context.Context
}

func newJoinRequest(ctx context.Context, user uuid.UUID, offer *webrtc.SessionDescription) *joinRequest {
	errChan := make(chan error)
	resChan := make(chan *joinResponse)

	return &joinRequest{
		offer:    offer,
		user:     user,
		err:      errChan,
		response: resChan,
		ctx:      ctx,
	}
}

type joinResponse struct {
	answer       *webrtc.SessionDescription
	resource     uuid.UUID
	RtpSessionId uuid.UUID
}

type leaveRequest struct {
	user     uuid.UUID
	response chan bool
	err      chan error
	ctx      context.Context
}
