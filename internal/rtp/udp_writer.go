package rtp

import (
	"errors"
	"fmt"
	"io"

	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

type udpWriter struct {
	id         string
	quit       chan struct{}
	globalQuit <-chan struct{}
}

func newUdpWriter(id string, globalQuit <-chan struct{}) *localWriter {
	return &localWriter{
		id:         id,
		quit:       make(chan struct{}),
		globalQuit: globalQuit,
	}
}

func (w *udpWriter) writeRtp(remoteTrack *webrtc.TrackRemote, localTrack *webrtc.TrackLocalStaticRTP) error {
	rtpBuf := make([]byte, rtpBufferSize)
	for {
		select {
		case <-w.globalQuit:
			slog.Info("rtp.localWriter closed globally", "track id", w.id)
			return nil
		case <-w.quit:
			slog.Info("rtp.localWriter closed locally", "track id", w.id)
			return nil
		default:
			i, _, err := remoteTrack.Read(rtpBuf)
			switch {
			case errors.Is(err, io.EOF):
				w.close()
				return nil
			case err != nil:
				w.close()
				return fmt.Errorf("reading rtp buffer: %w", err)
			}
			// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
			if _, err := localTrack.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				// stop reading because writing error
				w.close()
				return fmt.Errorf("reading rtp buffer: %w", err)
			}
		}
	}
}

func (w *udpWriter) close() {
	slog.Info("rtp.udpWriter: close", "track id", w.id)
	select {
	case <-w.globalQuit:
		slog.Warn("rtp.udpWriter the writer was already closed, con not close by global again", "track id", w.id)
	case <-w.quit:
		slog.Warn("rtp.udpWriter the Writer was already closed, con not close by local again", "track id", w.id)
	default:
		close(w.quit)
		slog.Info("rtp.udpWriter close was triggered", "track id", w.id)
	}
}
