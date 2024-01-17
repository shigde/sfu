package rtp

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/rtp/stats"
	"github.com/shigde/sfu/internal/telemetry"
	"go.opentelemetry.io/otel"
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
	trackInfos    map[string]*TrackInfo
	statsRegistry *stats.Registry
}

func newReceiver(sessionCxt context.Context, sessionId uuid.UUID, liveStream uuid.UUID, d TrackDispatcher, trackInfos map[string]*TrackInfo) *receiver {
	streams := make(map[string]*mediaStream)
	return &receiver{
		RWMutex:    sync.RWMutex{},
		id:         sessionId,
		sessionCxt: sessionCxt,
		liveStream: liveStream,
		streams:    streams,
		dispatcher: d,
		trackInfos: trackInfos,
	}
}

func (r *receiver) onTrack(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
	ctx, span := otel.Tracer(tracerName).Start(context.Background(), "rtp.receiver: on ingress track")
	defer span.End()

	stream := r.getStream(r.sessionCxt, r.id, remoteTrack.StreamID(), remoteTrack.ID())
	span.SetAttributes(attribute.String("streamId", remoteTrack.StreamID()))
	span.SetAttributes(attribute.String("track", remoteTrack.ID()))
	span.SetAttributes(attribute.String("kind", remoteTrack.Kind().String()))
	span.SetAttributes(attribute.String("purpose", stream.getPurpose().ToString()))
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
			telemetry.RecordError(span, err)
			// stop handler goroutine because error
			return
		}
		trackInfo = newTrackInfo(r.id, stream.getAudioTrack(), stream.getPurpose())
	}

	if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "video") {
		slog.Debug("rtp.receiver: on ingress video track", "streamId", remoteTrack.StreamID(), "track", remoteTrack.ID(), "kind", remoteTrack.Kind(), "purpose", stream.getPurpose().ToString())
		if err := stream.writeVideoRtp(ctx, remoteTrack, rtpReceiver); err != nil {
			slog.Error("rtp.receiver: on ingress video track", "err", err, "streamId", remoteTrack.StreamID(), "track", remoteTrack.ID(), "kind", remoteTrack.Kind(), "purpose", stream.getPurpose().ToString())
			telemetry.RecordError(span, err)
			// stop handler goroutine because error
			return
		}
		trackInfo = newTrackInfo(r.id, stream.getVideoTrack(), stream.getPurpose())
	}

	slog.Debug("rtp.receiver: info track", "streamId", trackInfo.GetTrackLocal().StreamID(), "track", trackInfo.GetTrackLocal().ID(), "kind", trackInfo.GetTrackLocal().Kind(), "purpose", trackInfo.Purpose.ToString())
	// send track to Lobby Hub
	r.dispatcher.DispatchAddTrack(trackInfo)

}
func (r *receiver) getTrackInfo(streamID string, trackId string) *TrackInfo {
	msid := fmt.Sprintf("%s %s", streamID, trackId)
	info, found := r.trackInfos[msid]
	if !found {
		info = &TrackInfo{SessionId: r.id, TrackSdpInfo: TrackSdpInfo{Purpose: PurposeGuest}}
		r.trackInfos[msid] = info
	}
	return info
}

func (r *receiver) getStream(sessionCxt context.Context, sessionId uuid.UUID, streamId string, trackId string) *mediaStream {
	r.Lock()
	defer r.Unlock()
	info := r.getTrackInfo(streamId, trackId)
	stream, ok := r.streams[streamId]
	if !ok {
		stream = newMediaStream(sessionCxt, streamId, sessionId, r.dispatcher, info.Purpose)
		r.streams[streamId] = stream
	}
	return stream
}
