package rtp

import (
	"fmt"
	"log"

	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v3"
)

type Engine struct {
	config webrtc.Configuration
	api    *webrtc.API
}

func NewEngine(rtpConfig *RtpConfig) (*Engine, error) {
	config := rtpConfig.getWebrtcConf()

	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, fmt.Errorf("register  default codecs: %w ", err)
	}

	// Create a InterceptorRegistry. This is the user configurable RTP/RTCP Pipeline.
	// This provides NACKs, RTCP Reports and other features. If you use `webrtc.NewPeerConnection`
	// this is enabled by default. If you are manually managing You MUST create a InterceptorRegistry
	// for each PeerConnection.
	i := &interceptor.Registry{}

	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		return nil, fmt.Errorf("register default interceptors: %w ", err)
	}

	// Register a intervalpli factory
	// This interceptor sends a PLI every 3 seconds. A PLI causes a video keyframe to be generated by the sender.
	// This makes our video seekable and more error resilent, but at a cost of lower picture quality and higher bitrates
	// A real world application should process incoming RTCP packets from viewers and forward them to senders
	intervalPliFactory, err := intervalpli.NewReceiverInterceptor()
	if err != nil {
		return nil, fmt.Errorf("create interval Pli factory: %w ", err)
	}
	i.Add(intervalPliFactory)

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i))
	return &Engine{
		config: config,
		api:    api,
	}, nil
}

func (e Engine) NewConnection(offer webrtc.SessionDescription) (*Connection, error) {
	peerConnection, err := e.api.NewPeerConnection(e.config)
	if err != nil {
		return nil, fmt.Errorf("create peer connection: %w ", err)
	}
	receiver := newReceiver()
	sender := newSender()

	peerConnection.OnTrack(receiver.onTrack)

	peerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
		if i == webrtc.ICEConnectionStateFailed {
			if err := peerConnection.Close(); err != nil {
				log.Println(err)
			}
			receiver.stop()
		}
	})

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		return nil, err
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)

	if err != nil {
		return nil, err
	}

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	return &Connection{
		peerConnector:  peerConnection,
		receiver:       receiver,
		sender:         sender,
		gatherComplete: gatherComplete,
	}, nil
}
