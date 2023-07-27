package lobby

import (
	"context"

	"github.com/pion/webrtc/v3"
)

type offerRequest struct {
	offerType
	offer  *webrtc.SessionDescription
	answer chan *webrtc.SessionDescription
	err    chan error
	ctx    context.Context
}

type offerType int

const (
	offerTypeReceving offerType = iota + 1
	offerTypeSending
)

func newOfferRequest(ctx context.Context, offer *webrtc.SessionDescription, offerType offerType) *offerRequest {
	return &offerRequest{
		offerType: offerType,
		offer:     offer,
		answer:    make(chan *webrtc.SessionDescription),
		err:       make(chan error),
		ctx:       ctx,
	}
}
