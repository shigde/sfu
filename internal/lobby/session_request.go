package lobby

import (
	"context"

	"github.com/pion/webrtc/v3"
)

type sessionRequest struct {
	sessionReqType
	reqSDP      *webrtc.SessionDescription
	respSDPChan chan *webrtc.SessionDescription
	err         chan error
	ctx         context.Context
}

type sessionReqType int

const (
	offerIngressReq sessionReqType = iota + 1
	initEgressReq
	answerEgressReq
	closeReq
	offerStaticEgressReq
	offerHostIngressReq
	offerHostEgressReq
	answerHostEgressReq
)

func newSessionRequest(ctx context.Context, sdp *webrtc.SessionDescription, reqType sessionReqType) *sessionRequest {
	return &sessionRequest{
		sessionReqType: reqType,
		reqSDP:         sdp,
		respSDPChan:    make(chan *webrtc.SessionDescription),
		err:            make(chan error),
		ctx:            ctx,
	}
}
