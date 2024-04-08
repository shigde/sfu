package rtp

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

type mediaWriter struct {
	id         string
	sessionCxt context.Context
	quit       chan struct{}
}

func newMediaWriter(sessionCxt context.Context, id string) *mediaWriter {
	return &mediaWriter{
		id:         id,
		sessionCxt: sessionCxt,
		quit:       make(chan struct{}),
	}
}

func (w *mediaWriter) writeRtp(remoteTrack *webrtc.TrackRemote, localTrack *webrtc.TrackLocalStaticRTP) error {
	rtpBuf := make([]byte, rtpBufferSize)
	slog.Debug("rtp.mediaWriter write RTP", "track id", w.id)
	for {
		select {
		case <-w.sessionCxt.Done():
			slog.Info("rtp.mediaWriter closed by session closed", "track id", w.id)
			return nil
		case <-w.quit:
			slog.Debug("rtp.mediaWriter closed locally", "track id", w.id)
			return nil
		default:
			i, _, err := remoteTrack.Read(rtpBuf)
			switch {
			case errors.Is(err, io.EOF):
				slog.Debug("rtp.mediaWriter closed EOF", "track id", w.id)
				w.close()
				return nil
			case err != nil:
				w.close()
				slog.Error("rtp.mediaWriter reading rtp buffer", "track id", w.id)
				return fmt.Errorf("reading rtp buffer: %w", err)
			}
			// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
			if _, err := localTrack.Write(rtpBuf[:i]); err != nil {
				// stop reading because writing error
				if !errors.Is(err, io.ErrClosedPipe) {
					w.close()
					slog.Warn("bug-1: rtp.mediaWriter err reading rtp buffer", "track id", w.id)
					return fmt.Errorf("bug-1: reading rtp buffer: %w", err)
				}
				slog.Debug("rtp.mediaWriter CLOSED PIPE!", "track id", w.id)
			}
		}
	}
}

func (w *mediaWriter) close() {
	slog.Info("rtp.mediaWriter: close", "track id", w.id)
	select {
	case <-w.sessionCxt.Done():
		slog.Warn("rtp.mediaWriter the mediaWriter was already closed, con not close again", "track id", w.id)
	case <-w.quit:
		slog.Warn("rtp.mediaWriter the mediaWriter was already closed, con not close by local again", "track id", w.id)
	default:
		close(w.quit)
		slog.Info("rtp.mediaWriter close was triggered", "track id", w.id)
	}
}

func (w *mediaWriter) isRunning() bool {
	select {
	case <-w.sessionCxt.Done():
		return false
	case <-w.quit:
		return false
	default:
		return true
	}
}
