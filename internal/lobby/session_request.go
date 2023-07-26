package lobby

import (
	"context"

	"github.com/pion/webrtc/v3"
)

type offerRequest struct {
	offer  *webrtc.SessionDescription
	answer chan *webrtc.SessionDescription
	err    chan error
	ctx    context.Context
}

func newOfferRequest(ctx context.Context, offer *webrtc.SessionDescription) *offerRequest {
	return &offerRequest{
		offer:  offer,
		answer: make(chan *webrtc.SessionDescription),
		err:    make(chan error),
		ctx:    ctx,
	}
}
