package client

import "github.com/pion/webrtc/v3"

type HostApi struct {
	*Client
	UserId  string
	Token   string
	ShigUrl string
}

func NewHostApi(token string, opt ...ClientOption) *HostApi {
	client := NewClient(opt...)
	return &HostApi{
		Client: client,
		Token:  token,
	}
}

func (a *HostApi) PostHostOffer(space string, stream string, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	return nil, nil
}
