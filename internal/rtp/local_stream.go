package rtp

import (
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

const rtpBufferSize = 1500

type localStream struct {
	id                     uuid.UUID
	remoteId               string
	audioTrack, videoTrack *webrtc.TrackLocalStaticRTP
	quit                   chan struct{}
}

func newLocalStream(remoteId string) *localStream {
	return &localStream{id: uuid.New(), remoteId: remoteId}
}

func (s *localStream) writeAudioRtp(track *webrtc.TrackRemote, dispatch chan *webrtc.TrackLocalStaticRTP) error {
	if err := s.addAudioTrack(track); err != nil {
		return fmt.Errorf("adding audio remote track (%s:%s) to local stream: %w", track.ID(), track.StreamID(), err)
	}

	// blocking until some reading
	// @TODO Fix me
	dispatch <- s.audioTrack
	return s.writeRtp(track, s.audioTrack)
}

func (s *localStream) addAudioTrack(remoteTrack *webrtc.TrackRemote) error {
	if s.audioTrack != nil {
		return errors.New("has already audio track")
	}
	audio, err := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "audio", s.id.String())
	if err != nil {
		return fmt.Errorf("creating new local audio track for local stream %s: %w", s.id, err)
	}
	s.audioTrack = audio
	return nil
}

func (s *localStream) writeVideoRtp(track *webrtc.TrackRemote, dispatch chan *webrtc.TrackLocalStaticRTP) error {
	if err := s.addAudioTrack(track); err != nil {
		return fmt.Errorf("adding video remote track (%s:%s) to local stream: %w", track.ID(), track.StreamID(), err)
	}

	// blocking until some reading
	// @TODO Fix me
	dispatch <- s.videoTrack
	return s.writeRtp(track, s.videoTrack)
}

func (s *localStream) addVideoTrack(remoteTrack *webrtc.TrackRemote) error {
	if s.videoTrack != nil {
		return errors.New("has already video track")
	}
	video, err := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "video", s.id.String())
	if err != nil {
		return fmt.Errorf("creating new local video track for local stream %s: %w", s.id, err)
	}
	s.videoTrack = video
	return nil
}

func (s *localStream) writeRtp(remoteTrack *webrtc.TrackRemote, localTrack *webrtc.TrackLocalStaticRTP) error {
	rtpBuf := make([]byte, rtpBufferSize)
	for {
		select {
		case <-s.quit:
			return nil
		default:
			i, _, err := remoteTrack.Read(rtpBuf)
			switch {
			case errors.Is(err, io.EOF):
				return nil
			case err != nil:
				return fmt.Errorf("reading rtp buffer: %w", err)
			}
			// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
			if _, err := localTrack.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				// stop reading because writing error
				return fmt.Errorf("reading rtp buffer: %w", err)
			}
		}
	}
}

func (s *localStream) close() {
	slog.Info("rtp.local_stream: close", "id", s.id)
	select {
	case <-s.quit:
		slog.Warn("rtp.local_stream: the stream was already closed", "id", s.id)
	default:
		close(s.quit)
		slog.Info("rtp.local_stream: close was triggered", "id", s.id)
	}
}
