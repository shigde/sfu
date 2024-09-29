package rtp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/dtls/v2"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/rtp/stats"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
)

var ErrIceGatheringInterruption = errors.New("getting ice gathering interrupted")
var ErrSessionClosed = errors.New("process interrupted because session closed")

type Endpoint struct {
	sessionCxt             context.Context
	sessionId              string
	liveStreamId           string
	endpointType           EndpointType
	peerConnection         peerConnection
	receiver               *receiver
	trackSdpInfoRepository *trackSdpInfoRepository
	// Means ice candidates are gathered
	// The ICE candidates are exchanged via the SDP, so the caller must wait until
	// all candidates have been written into the SDP.
	gatherComplete <-chan struct{}
	// Means the "Offer Answer Exchange" cycle is complete.
	// An egress endpoint is created without tracks. The renegotiation process begins
	// only the connection is established (initComplete) for the first time.
	initComplete  chan struct{}
	closed        chan struct{}
	statsRegistry *stats.Registry
	iceState      webrtc.ICEConnectionState
	// With Endpoint Optionals #######################################
	onChannel           func(dc *webrtc.DataChannel)
	onEstablished       func()
	onNegotiationNeeded func(offer webrtc.SessionDescription)
	waitBeforeONNSetup  <-chan struct{}
	onLostConnection    func()
	onIceStateConnected func()
	getCurrentTracksCbk func(ctx context.Context, sessionId uuid.UUID) ([]*TrackInfo, error)
	initTracks          []*initTrack // deprecated
	dispatcher          TrackDispatcher
}

func newEndpoint(sessionCxt context.Context, sessionId string, liveStreamId string, endpointType EndpointType, options ...EndpointOption) *Endpoint {
	endpoint := &Endpoint{
		sessionCxt:             sessionCxt,
		sessionId:              sessionId,
		liveStreamId:           liveStreamId,
		endpointType:           endpointType,
		trackSdpInfoRepository: newTrackSdpInfoRepository(),
		initComplete:           make(chan struct{}),
	}
	for _, opt := range options {
		opt(endpoint)
	}

	go func(ep *Endpoint) {
		select {
		case <-sessionCxt.Done():
			if err := ep.Destruct(); err != nil {
				slog.Error("rtp.endpoint: destruct endpoint", "err", err)
			}
		}
	}(endpoint)
	return endpoint
}

func (c *Endpoint) GetLocalDescription(ctx context.Context) (*webrtc.SessionDescription, error) {
	// block until ice gathering is complete before return local sdp
	// all ice candidates should be part of the answer
	_, span := rtpTrace(ctx, "endpoint_get_local_description")
	defer span.End()
	select {
	case <-c.gatherComplete:
		var err error
		offer := c.peerConnection.LocalDescription()
		if c.endpointType == EgressEndpoint {
			offer, err = setEgressTrackInfo(c.peerConnection.LocalDescription(), c.trackSdpInfoRepository)
			if err != nil {
				slog.Error("rtp.establish_egress:: sender doRenegotiation dc", "err", err)
				span.RecordError(err)
			}
			slog.Debug("#### GetLocalDescription", "offer", offer.SDP)
		}
		return offer, nil
	case <-c.sessionCxt.Done():
		span.RecordError(ErrSessionClosed)
		return nil, ErrSessionClosed
	case <-ctx.Done():
		span.RecordError(ErrIceGatheringInterruption)
		return nil, ErrIceGatheringInterruption
	}
}
func (c *Endpoint) SetAnswer(sdp *webrtc.SessionDescription) error {
	return c.peerConnection.SetRemoteDescription(*sdp)
}

func (c *Endpoint) SetNewOffer(sdp *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {

	err := getIngressTrackSdpInfo(*sdp, uuid.MustParse(c.sessionId), c.trackSdpInfoRepository)

	if err := c.peerConnection.SetRemoteDescription(*sdp); err != nil {
		return nil, fmt.Errorf("set new offer: %w", err)
	}

	answer, err := c.peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}
	if err = c.peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}
	return &answer, nil
}

func (c *Endpoint) SetInitComplete() {
	select {
	case <-c.initComplete:
	case <-c.sessionCxt.Done():
	default:
		close(c.initComplete)
		<-c.initComplete
	}
}

func (c *Endpoint) IsInitComplete() bool {
	select {
	case <-c.initComplete:
		return true
	case <-c.sessionCxt.Done():
		return false
	default:
		return false
	}
}

func (c *Endpoint) hasTrack(track webrtc.TrackLocal) bool {
	slog.Debug("rtp.connection: has Tracks")
	rtpSenderList := c.peerConnection.GetSenders()
	for _, rtpSender := range rtpSenderList {
		if rtpTrack := rtpSender.Track(); rtpTrack != nil {
			if rtpTrack.ID() == track.ID() {
				return true
			}
		}
	}
	return false
}

func (c *Endpoint) getSender(track webrtc.TrackLocal) (*webrtc.RTPSender, bool) {
	slog.Debug("rtp.connection: has Tracks")
	rtpSenderList := c.peerConnection.GetSenders()
	for _, rtpSender := range rtpSenderList {
		if rtpTrack := rtpSender.Track(); rtpTrack != nil {
			if rtpTrack.ID() == track.ID() {
				return rtpSender, true
			}
		}
	}
	return nil, false
}

func (c *Endpoint) AddTrack(ctx context.Context, info *TrackInfo) {
	_, span := rtpTrace(ctx, "endpoint_add_track")
	defer span.End()
	track := info.GetTrackLocal()
	purpose := info.Purpose
	slog.Debug("rtp.endpoint: add track", "streamId", track.StreamID(), "trackId", track.ID(), "kind", track.Kind(), "purpose", purpose)
	if has := c.hasTrack(track); !has {
		slog.Debug("rtp.endpoint: add track to connection", "streamId", track.StreamID(), "trackId", track.ID(), "kind", track.Kind(), "purpose", purpose, "signalState", c.peerConnection.SignalingState().String())
		var sender *webrtc.RTPSender
		var err error

		span.AddEvent("Add Track to Connection", trace.WithAttributes(
			attribute.String("localTrack", track.ID())),
		)

		if sender, err = c.peerConnection.AddTrack(track); err != nil {
			span.RecordError(err)
			slog.Error("rtp.endpoint: add track to connection", "err", err, "streamId", track.StreamID(), "trackId", track.ID(), "kind", track.Kind(), "purpose", purpose, "signalState", c.peerConnection.SignalingState().String())
			return
		}

		sdpTrack := info.TrackSdpInfo
		sdpTrack.EgressTrackId = sender.Track().ID()
		for _, transceiver := range c.peerConnection.GetTransceivers() {
			if tSender := transceiver.Sender(); tSender == sender {
				sdpTrack.EgressMid = transceiver.Mid()
				break
			}
		}
		c.trackSdpInfoRepository.Set(info.Id, &sdpTrack)

		// collect stats
		if c.statsRegistry != nil {
			labels := metric.Labels{
				metric.Stream:       c.liveStreamId,
				metric.MediaStream:  track.StreamID(),
				metric.TrackId:      track.ID(),
				metric.TrackKind:    track.Kind().String(),
				metric.TrackPurpose: purpose.ToString(),
				metric.Direction:    c.endpointType.ToString(),
			}
			for _, param := range sender.GetParameters().Encodings {
				if err = c.statsRegistry.StartWorker(labels, param.SSRC); err != nil {
					slog.Error("rtp.endpoint: start stats worker", "err", err, "ssrc", param.SSRC)
				}
			}
		}
	}
}

func (c *Endpoint) RemoveTrack(ctx context.Context, info *TrackInfo) {
	_, span := rtpTrace(ctx, "endpoint_remove_track")
	defer span.End()
	track := info.GetTrackLocal()
	slog.Debug("rtp.endpoint: remove track", "streamId", track.StreamID(), "trackId", track.ID(), "purpose", track.Kind())

	if sender, has := c.getSender(track); has {
		c.trackSdpInfoRepository.Delete(info.GetId())

		span.AddEvent("Remove Track from Connection", trace.WithAttributes(
			attribute.String("localTrack", track.ID())),
		)
		slog.Debug("rtp.endpoint: remove track from connection", "streamId", track.StreamID(), "trackId", track.ID(), "kind", track.Kind())
		if err := c.peerConnection.RemoveTrack(sender); err != nil {
			span.RecordError(err)
			slog.Error("rtp.endpoint: remove track from connection", "err", err, "streamId", track.StreamID(), "trackId", track.ID(), "purpose", track.Kind())
		}

		if c.statsRegistry != nil {
			for _, param := range sender.GetParameters().Encodings {
				c.statsRegistry.StopWorker(param.SSRC)
			}
		}
	}
}

func (c *Endpoint) SetIngressMute(ingressMid string, mute bool) (*TrackInfo, bool) {
	if sdpInfo, ok := c.trackSdpInfoRepository.getTrackSdpInfoByIngressMid(ingressMid); ok {
		sdpInfo.Mute = mute
		return newTrackInfo(nil, *sdpInfo), true
	}
	return nil, false
}

func (c *Endpoint) SetEgressMute(infoId uuid.UUID, mute bool) (*TrackInfo, bool) {
	if sdpInfo, ok := c.trackSdpInfoRepository.Get(infoId); ok {
		sdpInfo.Mute = mute
		return newTrackInfo(nil, *sdpInfo), true
	}
	return nil, false
}

func (c *Endpoint) doRenegotiation() {
	if c.onNegotiationNeeded == nil {
		return
	}
	select {
	case <-c.sessionCxt.Done():
		return
	case <-c.initComplete:
	}

	slog.Debug("rtp.establish_egress: sender OnNegotiationNeeded was triggered")

	offer, err := c.peerConnection.CreateOffer(nil)
	if err != nil {
		slog.Error("rtp.establish_egress:: sender doRenegotiation", "err", err)
		return
	}
	gg := webrtc.GatheringCompletePromise(c.peerConnection.(*webrtc.PeerConnection))
	_ = c.peerConnection.SetLocalDescription(offer)
	select {
	case <-c.sessionCxt.Done():
		return
	case <-gg:
	}

	// munge sdp
	mungedOffer, err := setEgressTrackInfo(c.peerConnection.LocalDescription(), c.trackSdpInfoRepository)
	if err != nil {
		slog.Error("rtp.establish_egress:: sender doRenegotiation dc", "err", err)
	}
	slog.Debug("#### NegotiationNeeded", "offer", mungedOffer.SDP)
	c.onNegotiationNeeded(*mungedOffer)
}

func (c *Endpoint) onICEConnectionStateChange(state webrtc.ICEConnectionState) {
	slog.Debug("rtp.endpoint: ice state:", "state", state, "sessionId", c.sessionId, "type", c.endpointType)

	if state == webrtc.ICEConnectionStateFailed {
		slog.Warn("rtp.endpoint: endpoint become idle", "sessionId", c.sessionId, "type", c.endpointType)
		return
	}

	if state == webrtc.ICEConnectionStateDisconnected || state == webrtc.ICEConnectionStateClosed {
		slog.Warn("rtp.endpoint: lost connection:", "state", state, "sessionId", c.sessionId, "type", c.endpointType)
		if c.onLostConnection != nil {
			c.onLostConnection()
		}
	}

	if state == webrtc.ICEConnectionStateConnected && c.onIceStateConnected != nil {
		c.onIceStateConnected()
	}
}

func (c *Endpoint) Destruct() error {
	if c.statsRegistry != nil {
		c.statsRegistry.StopAllWorker()
	}

	if c.sessionId != "" && c.liveStreamId != "" {
		metric.GraphNodeDelete(metric.BuildNode(c.sessionId, c.liveStreamId, c.endpointType.ToString()))
	}

	if c.peerConnection == nil {
		return nil
	}

	for _, rtpSender := range c.peerConnection.GetSenders() {
		_ = rtpSender.Stop()
	}

	if err := c.peerConnection.Close(); err != nil && !errors.Is(err, dtls.ErrConnClosed) {
		return fmt.Errorf("closing peer connection: %w", err)
	}

	return nil
}

func (c *Endpoint) getPeerConnection() *webrtc.PeerConnection {
	pc, _ := c.peerConnection.(*webrtc.PeerConnection)
	return pc
}

type peerConnection interface {
	LocalDescription() *webrtc.SessionDescription
	SetLocalDescription(desc webrtc.SessionDescription) error
	SetRemoteDescription(desc webrtc.SessionDescription) error
	GetSenders() (result []*webrtc.RTPSender)
	GetTransceivers() []*webrtc.RTPTransceiver
	AddTrack(track webrtc.TrackLocal) (*webrtc.RTPSender, error)
	RemoveTrack(sender *webrtc.RTPSender) error
	OnTrack(f func(*webrtc.TrackRemote, *webrtc.RTPReceiver))
	SignalingState() webrtc.SignalingState
	CreateOffer(options *webrtc.OfferOptions) (webrtc.SessionDescription, error)
	CreateAnswer(options *webrtc.AnswerOptions) (webrtc.SessionDescription, error)
	OnICEConnectionStateChange(f func(webrtc.ICEConnectionState))
	OnNegotiationNeeded(f func())
	OnDataChannel(func(*webrtc.DataChannel))
	Close() error
}

type initTrack struct {
	purpose Purpose
	track   webrtc.TrackLocal
}

type EndpointType int

const (
	IngressEndpoint EndpointType = iota + 1
	EgressEndpoint
)

func (et EndpointType) ToString() string {
	switch et {
	case IngressEndpoint:
		return "ingress"
	case EgressEndpoint:
		return "egress"
	default:
		return "unknown"
	}
}
