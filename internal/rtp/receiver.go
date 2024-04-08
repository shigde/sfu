package rtp

import (
	"context"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/rtp/stats"
	"github.com/shigde/sfu/internal/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/exp/slog"
)

type receiver struct {
	sync.RWMutex
	id            uuid.UUID // session ID
	sessionCxt    context.Context
	liveStream    uuid.UUID
	streams       map[string]*mediaStream
	dispatcher    TrackDispatcher
	trackSdpInfos *trackSdpInfoRepository
	statsRegistry *stats.Registry
}

func newReceiver(sessionCxt context.Context, sessionId uuid.UUID, liveStream uuid.UUID, d TrackDispatcher, trackSdpInfos *trackSdpInfoRepository) *receiver {
	streams := make(map[string]*mediaStream)
	return &receiver{
		RWMutex:       sync.RWMutex{},
		id:            sessionId,
		sessionCxt:    sessionCxt,
		liveStream:    liveStream,
		streams:       streams,
		dispatcher:    d,
		trackSdpInfos: trackSdpInfos,
	}
}

func (r *receiver) onTrack(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
	ctx, span := newTraceSpan(context.Background(), r.sessionCxt, "rtp.ingress: add_"+remoteTrack.Kind().String()+"_track")
	defer span.End()

	trackSdpInfo := r.getIngressTrackSdpInfo(remoteTrack.ID())
	trackSdpInfo.IngressMid = rtpReceiver.RTPTransceiver().Mid()

	stream := r.getStream(r.sessionCxt, r.id, remoteTrack.StreamID(), *trackSdpInfo, remoteTrack.Kind().String())
	span.SetAttributes(
		attribute.String("mediaStreamId", remoteTrack.StreamID()),
		attribute.String("track", remoteTrack.ID()),
		attribute.String("kind", remoteTrack.Kind().String()),
		attribute.String("purpose", stream.getPurpose().ToString()),
	)
	slog.Debug("rtp.receiver: on ingress track", "streamId", remoteTrack.StreamID(), "track", remoteTrack.ID(), "kind", remoteTrack.Kind(), "purpose", stream.getPurpose().ToString())

	// collect metrics
	if r.statsRegistry != nil {
		labels := metric.Labels{
			metric.Stream:       r.liveStream.String(),
			metric.MediaStream:  remoteTrack.StreamID(),
			metric.TrackId:      remoteTrack.ID(),
			metric.TrackKind:    remoteTrack.Kind().String(),
			metric.TrackPurpose: stream.getPurpose().ToString(),
			metric.Direction:    IngressEndpoint.ToString(),
		}
		if err := r.statsRegistry.StartWorker(labels, remoteTrack.SSRC()); err != nil {
			slog.Error("rtp.receiver: start stats worker", "err", err)
		}
	}

	var trackInfo *TrackInfo
	if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "audio") {
		slog.Debug("rtp.receiver: on ingress audio track", "streamId", remoteTrack.StreamID(), "track", remoteTrack.ID(), "kind", remoteTrack.Kind(), "purpose", stream.getPurpose().ToString())
		if err := stream.writeAudioRtp(ctx, remoteTrack, rtpReceiver); err != nil {
			slog.Error("rtp.receiver: on ingress audio track", "err", err, "streamId", remoteTrack.StreamID(), "track", remoteTrack.ID(), "kind", remoteTrack.Kind(), "purpose", stream.getPurpose().ToString())
			_ = telemetry.RecordError(span, err)
			// stop handler goroutine because error
			return
		}
		trackInfo = newTrackInfo(stream.getAudioTrack(), *trackSdpInfo)
	}

	if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "video") {
		slog.Debug("rtp.receiver: on ingress video track", "streamId", remoteTrack.StreamID(), "track", remoteTrack.ID(), "kind", remoteTrack.Kind(), "purpose", stream.getPurpose().ToString())
		if err := stream.writeVideoRtp(ctx, remoteTrack, rtpReceiver); err != nil {
			slog.Error("rtp.receiver: on ingress video track", "err", err, "streamId", remoteTrack.StreamID(), "track", remoteTrack.ID(), "kind", remoteTrack.Kind(), "purpose", stream.getPurpose().ToString())
			_ = telemetry.RecordError(span, err)
			// stop handler goroutine because error
			return
		}

		trackInfo = newTrackInfo(stream.getVideoTrack(), *trackSdpInfo)
	}

	slog.Debug("rtp.receiver: info track", "streamId", trackInfo.GetTrackLocal().StreamID(), "track", trackInfo.GetTrackLocal().ID(), "kind", trackInfo.GetTrackLocal().Kind(), "purpose", trackInfo.Purpose.ToString())
	// send track to Lobby Hub
	r.dispatcher.DispatchAddTrack(ctx, trackInfo)

}
func (r *receiver) getIngressTrackSdpInfo(ingressTrackId string) *TrackSdpInfo {
	info, found := r.trackSdpInfos.getSdpInfoByIngressTrackId(ingressTrackId)
	if !found {
		info = newTrackSdpInfo(r.id)
		info.Purpose = PurposeGuest
		r.trackSdpInfos.Set(info.Id, info)
	}
	return info
}

func (r *receiver) getStream(sessionCxt context.Context, sessionId uuid.UUID, streamId string, sdpInfo TrackSdpInfo, kind string) *mediaStream {
	r.Lock()
	defer r.Unlock()
	stream, ok := r.streams[streamId]
	if !ok {
		stream = newMediaStream(sessionCxt, streamId, sessionId, r.dispatcher, sdpInfo.Purpose)
		r.streams[streamId] = stream
	}

	if kind == "video" {
		stream.setVideoSdpInfo(sdpInfo)
	}
	if kind == "audio" {
		stream.setAudioSdpInfo(sdpInfo)
	}

	return stream
}
