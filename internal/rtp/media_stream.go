package rtp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
)

const rtpBufferSize = 1500

type mediaStream struct {
	id                       uuid.UUID
	sessionCxt               context.Context
	remoteId                 string
	sessionId                uuid.UUID
	audioTrack, videoTrack   *webrtc.TrackLocalStaticRTP
	audioWriter, videoWriter *mediaWriter
	purpose                  Purpose
	dispatcher               TrackDispatcher
}

func newMediaStream(sessionCxt context.Context, remoteId string, sessionId uuid.UUID, dispatcher TrackDispatcher, purpose Purpose) *mediaStream {
	return &mediaStream{
		id:         uuid.New(),
		sessionCxt: sessionCxt,
		remoteId:   remoteId,
		sessionId:  sessionId,
		dispatcher: dispatcher,
		purpose:    purpose,
	}
}

func (s *mediaStream) writeAudioRtp(ctx context.Context, track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) error {
	slog.Debug("rtp.receiver: on track")
	ctx, span := otel.Tracer(tracerName).Start(ctx, "mediaStream:writeAudioRtp")
	defer span.End()
	audio, err := s.createNewAudioLocalTrack(track)
	if err != nil {
		return fmt.Errorf("adding audio remote track (%s:%s) to local stream: %w", track.ID(), track.StreamID(), err)
	}
	s.audioTrack = audio
	s.audioWriter = newMediaWriter(s.sessionCxt, s.audioTrack.ID())

	// start local audio track
	go func() {
		defer s.dispatcher.DispatchRemoveTrack(newTrackInfo(s.sessionId, s.getAudioTrack(), s.purpose))

		if err = s.audioWriter.writeRtp(track, s.audioTrack); err != nil {
			slog.Error("rtp.local_stream: writing local audio track ", "streamId", s.id, "err", err)
		}
		slog.Debug("rtp.local_stream: stop writing local audio track ", "streamId", s.id, "trackId", s.audioTrack.ID())
	}()
	return nil
}

func (s *mediaStream) writeVideoRtp(ctx context.Context, track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) error {
	_, span := otel.Tracer(tracerName).Start(ctx, "mediaStream:writeVideoRtp")
	defer span.End()

	video, err := s.createNewVideoLocalTrack(track)
	if err != nil {
		return fmt.Errorf("adding video remote track (%s:%s) to local stream: %w", track.ID(), track.StreamID(), err)
	}

	s.videoTrack = video
	s.videoWriter = newMediaWriter(s.sessionCxt, s.videoTrack.ID())

	// start local video track
	go func() {
		defer s.dispatcher.DispatchRemoveTrack(newTrackInfo(s.sessionId, s.getVideoTrack(), s.purpose))

		if err = s.videoWriter.writeRtp(track, s.videoTrack); err != nil {
			slog.Error("rtp.local_stream: writing local video track ", "streamId", s.id, "err", err)
		}
		slog.Debug("rtp.local_stream: stop writing local video track ", "streamId", s.id, "trackId", s.videoTrack.ID())
	}()

	return nil
}

func (s *mediaStream) createNewAudioLocalTrack(remoteTrack *webrtc.TrackRemote) (*webrtc.TrackLocalStaticRTP, error) {
	if s.audioTrack != nil {
		return nil, errors.New("has already audio track")
	}
	audio, err := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, uuid.NewString(), s.id.String())
	if err != nil {
		return nil, fmt.Errorf("creating new local audio track for local stream %s: %w", s.id, err)
	}
	return audio, nil
}

func (s *mediaStream) createNewVideoLocalTrack(remoteTrack *webrtc.TrackRemote) (*webrtc.TrackLocalStaticRTP, error) {
	if s.videoTrack != nil {
		return nil, errors.New("has already video track")
	}
	video, err := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, uuid.NewString(), s.id.String())
	if err != nil {
		return nil, fmt.Errorf("creating new local video track for local stream %s: %w", s.id, err)
	}
	return video, nil
}

func (s *mediaStream) getVideoTrack() *webrtc.TrackLocalStaticRTP {
	return s.videoTrack
}

func (s *mediaStream) getAudioTrack() *webrtc.TrackLocalStaticRTP {
	return s.audioTrack
}

func (s *mediaStream) getPurpose() Purpose {
	return s.purpose
}
