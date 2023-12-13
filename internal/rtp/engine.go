package rtp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/static"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
)

const tracerName = "github.com/shigde/sfu/internal/engine"

type Engine struct {
	config webrtc.Configuration
}

func NewEngine(rtpConfig *RtpConfig) (*Engine, error) {
	config := rtpConfig.getWebrtcConf()
	return &Engine{
		config: config,
	}, nil
}

func (e *Engine) createApi() (*webrtc.API, error) {
	im := newInterceptorMap()
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, fmt.Errorf("register  default codecs: %w ", err)
	}

	statsInterceptorFactory, err := stats.NewInterceptor()
	if err != nil {
		return nil, fmt.Errorf("create stats interceptor factory: %w", err)
	}

	go func() {
		statsInterceptorFactory.OnNewPeerConnection(func(id string, getter stats.Getter) {
			im.setStatsGetter(id, getter)
		})
	}()

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
	i.Add(statsInterceptorFactory)

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i))
	return api, nil
}

func (e *Engine) EstablishEgressEndpoint(ctx context.Context, sessionId uuid.UUID, offer webrtc.SessionDescription, dispatcher TrackDispatcher, handler StateEventHandler) (*Endpoint, error) {
	_, span := otel.Tracer(tracerName).Start(ctx, "engine:create egress endpoint")
	defer span.End()
	api, err := e.createApi()
	if err != nil {
		return nil, fmt.Errorf("creating api: %w", err)
	}

	peerConnection, err := api.NewPeerConnection(e.config)
	if err != nil {
		return nil, fmt.Errorf("create egress peer connection: %w ", err)
	}

	trackInfos, err := getTrackInfo(offer, sessionId)
	if err != nil {
		return nil, fmt.Errorf("parsing track info: %w ", err)
	}

	receiver := newReceiver(sessionId, dispatcher, nil, trackInfos)
	peerConnection.OnTrack(receiver.onTrack)

	peerConnection.OnICEConnectionStateChange(handler.OnConnectionStateChange)

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		slog.Debug("rtp.engine: egress endpoint new DataChannel", "label", d.Label(), "id", d.ID())
		handler.OnChannel(d)
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

	return &Endpoint{
		peerConnection: peerConnection,
		receiver:       receiver,
		gatherComplete: gatherComplete,
	}, nil
}

func (e *Engine) EstablishIngressEndpoint(ctx context.Context, sessionId uuid.UUID, sendingTracks []webrtc.TrackLocal, handler StateEventHandler) (*Endpoint, error) {
	_, span := otel.Tracer(tracerName).Start(ctx, "engine:create ingress endpoint")
	defer span.End()
	api, err := e.createApi()
	if err != nil {
		return nil, fmt.Errorf("creating api: %w", err)
	}

	peerConnection, err := api.NewPeerConnection(e.config)
	if err != nil {
		return nil, fmt.Errorf("create sender peer connection: %w ", err)
	}

	peerConnection.OnICEConnectionStateChange(handler.OnConnectionStateChange)

	initComplete := make(chan struct{})

	// @TODO: Fix the race
	// First we create the sender endpoint and after this we add the individual tracks.
	// I don't know why, but Pion doesn't trigger renegotiation when creating a peer connection with tracks and the sdp
	// exchange is not finish. A peer connection without tracks where all tracks are added afterwards triggers renegotiation.
	// Unfortunately, "sendingTracks" could be outdated in the meantime.
	// This creates a race between remove and add track that I still have to think about it.
	go func() {
		<-initComplete
		if sendingTracks != nil {
			for _, track := range sendingTracks {
				if _, err = peerConnection.AddTrack(track); err != nil {
					slog.Error("rtp.engine: adding track to connection", "err", err)
				}
			}
		}
	}()

	peerConnection.OnNegotiationNeeded(func() {
		<-initComplete
		slog.Debug("rtp.engine: sender OnNegotiationNeeded was triggered")
		offer, err := peerConnection.CreateOffer(nil)
		if err != nil {
			slog.Error("rtp.engine: sender OnNegotiationNeeded", "err", err)
			return
		}
		gg := webrtc.GatheringCompletePromise(peerConnection)
		_ = peerConnection.SetLocalDescription(offer)
		<-gg
		handler.OnNegotiationNeeded(*peerConnection.LocalDescription())
	})
	slog.Debug("rtp.engine: sender: OnNegotiationNeeded setup finish")

	err = creatDC(peerConnection, handler)
	if err != nil {
		return nil, fmt.Errorf("creating data channel: %w", err)
	}

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("creating offer: %w", err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	if err = peerConnection.SetLocalDescription(offer); err != nil {
		return nil, err
	}

	return &Endpoint{
		peerConnection: peerConnection,
		gatherComplete: gatherComplete,
		initComplete:   initComplete,
	}, nil
}
func (e *Engine) EstablishStaticEgressEndpoint(ctx context.Context, sessionId uuid.UUID, offer webrtc.SessionDescription, options ...EndpointOption) (*Endpoint, error) {
	_, span := otel.Tracer(tracerName).Start(ctx, "engine:create static egress endpoint")
	defer span.End()
	api, err := e.createApi()
	if err != nil {
		return nil, fmt.Errorf("creating api: %w", err)
	}

	peerConnection, err := api.NewPeerConnection(e.config)
	if err != nil {
		return nil, fmt.Errorf("create receiver peer connection: %w ", err)
	}
	endpoint := &Endpoint{
		peerConnection: peerConnection,
	}
	for _, opt := range options {
		opt(endpoint)
	}

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		return nil, err
	}

	endpoint.gatherComplete = webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	return endpoint, nil
}

func creatDC(pc *webrtc.PeerConnection, handler StateEventHandler) error {
	ordered := false
	maxRetransmits := uint16(0)

	options := &webrtc.DataChannelInit{
		Ordered:        &ordered,
		MaxRetransmits: &maxRetransmits,
	}

	// Create a datachannel with label 'data'
	dc, err := pc.CreateDataChannel("data", options)
	if err != nil {
		return fmt.Errorf("creating data channel: %w", err)
	}
	handler.OnChannel(dc)
	return nil
}

// NewStaticMediaSenderEndpoint can be used to send static streams from file in a lobby.
// @deprecated
func (e *Engine) NewStaticMediaSenderEndpoint(media *static.MediaFile) (*Endpoint, error) {
	stateHandler := newMediaStateEventHandler()
	api, err := e.createApi()
	if err != nil {
		return nil, fmt.Errorf("creating api: %w", err)
	}
	peerConnection, err := api.NewPeerConnection(e.config)
	if err != nil {
		return nil, fmt.Errorf("create receiver peer connection: %w ", err)
	}

	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		slog.Debug("rtp.engine: connection State has changed", "state", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

	rtpVideoSender, err := peerConnection.AddTrack(media.VideoTrack)
	if err != nil {
		return nil, fmt.Errorf("add video track to peer connection: %w ", err)
	}
	media.PlayVideo(iceConnectedCtx, rtpVideoSender)

	rtpAudioSender, err := peerConnection.AddTrack(media.AudioTrack)
	if err != nil {
		return nil, fmt.Errorf("add audio track to peer connection: %w ", err)
	}
	media.PlayAudio(iceConnectedCtx, rtpAudioSender)

	err = creatDC(peerConnection, stateHandler)

	if err != nil {
		return nil, fmt.Errorf("creating data channel: %w", err)
	}

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("creating offer: %w", err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	if err = peerConnection.SetLocalDescription(offer); err != nil {
		return nil, err
	}

	return &Endpoint{peerConnection: peerConnection, gatherComplete: gatherComplete}, nil
}

// NewSignalConnection can be used to listen on lobby events.
func (e *Engine) NewSignalConnection(ctx context.Context, handler StateEventHandler) (*Connetcion, error) {
	api, err := e.createApi()
	if err != nil {
		return nil, fmt.Errorf("creating api: %w", err)
	}
	peerConnection, err := api.NewPeerConnection(e.config)
	if err != nil {
		return nil, fmt.Errorf("create receiver peer connection: %w ", err)
	}

	_, iceConnectedCtxCancel := context.WithCancel(ctx)

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		// fmt.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

	err = creatDC(peerConnection, handler)

	if err != nil {
		return nil, fmt.Errorf("creating data channel: %w", err)
	}

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("creating offer: %w", err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	if err = peerConnection.SetLocalDescription(offer); err != nil {
		return nil, err
	}

	return &Connetcion{PeerConnection: peerConnection, GatherComplete: gatherComplete}, nil
}

// NewStaticReceiverEndpoint can be used to receive Medias from a lobby.
func (e *Engine) NewReceiverConnection(ctx context.Context, offer webrtc.SessionDescription, handler StateEventHandler, rtmpEndpoint string) (*Connetcion, error) {
	api, err := e.createApi()
	if err != nil {
		return nil, fmt.Errorf("creating api: %w", err)
	}
	peerConnection, err := api.NewPeerConnection(e.config)
	if err != nil {
		return nil, fmt.Errorf("create receiver peer connection: %w ", err)
	}

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		slog.Debug("rtp.engine: receiverEndpoint new DataChannel", "label", d.Label(), "id", d.ID())
		handler.OnChannel(d)
	})
	// Allow us to receive 1 audio track, and 1 video track
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		slog.Error("rtp.engine: .addTransceiverFromKind audio", "err", err)
	} else if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		slog.Error("rtp.engine: .addTransceiverFromKind video", "err", err)
	}

	go func(ctx context.Context, pc *webrtc.PeerConnection, rtmp string) {
		rtmpListener(ctx, pc, rtmp)
	}(ctx, peerConnection, rtmpEndpoint)

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

	return &Connetcion{
		PeerConnection: peerConnection,
		GatherComplete: gatherComplete,
	}, nil
}
