package rtp

import (
	"errors"
	"net"

	"github.com/pion/rtp"
	"golang.org/x/exp/slog"
)

type udpWriter struct {
	id         string
	udp        *UdpConnection
	quit       chan struct{}
	globalQuit <-chan struct{}
}

func newUdpWriter(id string, udp *UdpConnection, globalQuit <-chan struct{}) *udpWriter {
	return &udpWriter{
		id:         id,
		udp:        udp,
		quit:       make(chan struct{}),
		globalQuit: globalQuit,
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

func (w *udpWriter) WriteRTP(header *rtp.Header, payload []byte) (int, error) {
	pkt, err := (&rtp.Packet{Header: *header, Payload: payload}).Marshal()
	if err != nil {
		return 0, err
	}

	n, writeErr := w.udp.conn.Write(pkt)
	if writeErr != nil {
		// For this particular example, third party applications usually timeout after a short
		// amount of time during which the user doesn't have enough time to provide the answer
		// to the browser.
		// That's why, for this particular example, the user first needs to provide the answer
		// to the browser then open the third party application. Therefore we must not kill
		// the forward on "connection refused" errors
		var opError *net.OpError
		if errors.As(writeErr, &opError) && opError.Err.Error() == "write: connection refused" {
			return n, nil
		}
		return n, writeErr
	}
	return n, nil
}

// Write encrypts and writes a full RTP packet
func (w *udpWriter) Write(b []byte) (int, error) {
	return 0, nil
}
