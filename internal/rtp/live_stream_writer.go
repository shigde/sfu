package rtp

import (
	"errors"
	"net"
	"sync"

	"github.com/pion/rtp"
	"golang.org/x/exp/slog"
)

type liveStreamWriter struct {
	sync.Mutex
	id         string
	udp        *UdpConnection
	quit       chan struct{}
	globalQuit <-chan struct{}
}

func newLiveStreamWriter(id string, udp *UdpConnection, globalQuit <-chan struct{}) *liveStreamWriter {
	return &liveStreamWriter{
		id:         id,
		udp:        udp,
		quit:       make(chan struct{}),
		globalQuit: globalQuit,
	}
}

func (w *liveStreamWriter) close() {
	slog.Info("rtp.liveStreamWriter: close", "track id", w.id)
	select {
	case <-w.globalQuit:
		slog.Warn("rtp.liveStreamWriter the writer was already closed, con not close by global again", "track id", w.id)
	case <-w.quit:
		slog.Warn("rtp.liveStreamWriter the Writer was already closed, con not close by local again", "track id", w.id)
	default:
		close(w.quit)
		slog.Info("rtp.liveStreamWriter close was triggered", "track id", w.id)
	}
}

// bufferpool is a global pool of buffers used for encrypted packets in
// writeRTP below.  Since it's global, buffers can be shared between
// different sessions, which amortizes the cost of allocating the pool.
//
// 1472 is the maximum Ethernet UDP payload.  We give ourselves 20 bytes
// of slack for any authentication tags, which is more than enough for
// either CTR or GCM.  If the buffer is too small, no harm, it will just
// get expanded by growBuffer.
var bufferpool = sync.Pool{ // nolint:gochecknoglobals
	New: func() interface{} {
		return make([]byte, 1492)
	},
}

func (w *liveStreamWriter) WriteRTP(header *rtp.Header, payload []byte) (int, error) {

	// encryptRTP will either return our buffer, or, if it is too
	// small, allocate a new buffer itself.  In either case, it is
	// safe to put the buffer back into the pool, but only after
	// nextConn.Write has returned.
	//ibuf := bufferpool.Get()
	//defer bufferpool.Put(ibuf)
	//w.Lock()
	//encrypted, err := w.encryptRTP(ibuf.([]byte), header, payload)
	//w.Unlock()

	pkg, err := (&rtp.Packet{Header: *header, Payload: payload}).Marshal()
	if err != nil {
		return 0, err
	}

	n, writeErr := w.udp.conn.Write(pkg)
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

func (w *liveStreamWriter) encryptRTP(dst []byte, header *rtp.Header, payload []byte) (ciphertext []byte, err error) {
	dst = growBufferSize(dst, header.MarshalSize()+len(payload))
	_, err = (&rtp.Packet{Header: *header, Payload: payload}).MarshalTo(dst)
	return dst, err

}

// Write encrypts and writes a full RTP packet
func (w *liveStreamWriter) Write(b []byte) (int, error) {
	return 0, nil
}

func growBufferSize(buf []byte, size int) []byte {
	if size <= cap(buf) {
		return buf[:size]
	}

	buf2 := make([]byte, size)
	copy(buf2, buf)
	return buf2
}
