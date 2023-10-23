package rtp

import (
	"errors"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

type UdpForwarder struct {
	id           uuid.UUID
	audio, video *UdpConnection
	quit         chan struct{}
}

func NewUdpForwarder(id uuid.UUID) (*UdpForwarder, error) {
	quit := make(chan struct{})

	f := &UdpForwarder{
		id:    id,
		audio: &UdpConnection{port: 4000, payloadType: 111},
		video: &UdpConnection{port: 4002, payloadType: 96},
		quit:  quit,
	}

	go f.Run()
	return f, nil
}

func (f *UdpForwarder) Run() {
	defer func() {
		slog.Info("forwarder stop running", "forwarderID", f.id)
		if err := f.close(); err != nil {
			slog.Error("forwarder closing udp ports", "err", err, "forwarderID", f.id)
		}
	}()

	var err error
	var laddr *net.UDPAddr
	if laddr, err = net.ResolveUDPAddr("udp", "127.0.0.1:"); err != nil {
		slog.Error("forwarder: resolving udp address", "err", err, "forwarderID", f.id.String())
		return
	}

	if err = f.connect(f.audio, laddr); err != nil {
		slog.Error("forwarder: connecting audio", "err", err, "forwarderID", f.id.String())
		return
	}
	if err = f.connect(f.video, laddr); err != nil {
		slog.Error("forwarder: connecting video", "err", err, "forwarderID", f.id.String())
		return
	}

	select {
	case <-f.quit:
	}
}

func (f *UdpForwarder) Stop() {
	select {
	case <-f.quit:
		return
	default:
		close(f.quit)
		<-f.quit
	}
}

func (f *UdpForwarder) AddTrack(track *TrackInfo) {
	if track.RemoteTrack.Kind() == webrtc.RTPCodecTypeAudio {
		go func() {
			slog.Info("forwarder: writing audio", "forwarderId", f.id)
			if err := f.writeTrack(f.audio, track.RemoteTrack); err != nil {
				slog.Error("writing stream audio", "err", err, "forwarderId", f.id)
			}
			slog.Info("stop writing stream audio", "forwarderId", f.id)
		}()
	}

	if track.RemoteTrack.Kind() == webrtc.RTPCodecTypeVideo {
		go func() {
			slog.Info("forwarder: writing video", "forwarderId", f.id)
			if err := f.writeTrack(f.video, track.RemoteTrack); err != nil {
				slog.Error("writing stream video", "err", err, "forwarderId", f.id)
			}
			slog.Info("forwarder: stop writing stream video", "forwarderId", f.id)
		}()
	}
}

func (f *UdpForwarder) connect(udp *UdpConnection, laddr *net.UDPAddr) error {
	// Create remote addr
	var raddr *net.UDPAddr
	var err error
	if raddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", udp.port)); err != nil {
		return fmt.Errorf("resolving udp port: %w", err)
	}

	// Dial udp
	if udp.conn, err = net.DialUDP("udp", laddr, raddr); err != nil {
		return fmt.Errorf("dealing udp port: %w", err)
	}
	return err
}

func (f *UdpForwarder) close() error {
	if f.audio.conn != nil {
		if err := f.audio.conn.Close(); err != nil {
			return fmt.Errorf("closing audio udp port: %w", err)
		}
	}
	if f.video.conn != nil {
		if err := f.video.conn.Close(); err != nil {
			return fmt.Errorf("closing video udp port: %w", err)
		}
	}
	return nil
}

func (f *UdpForwarder) writeTrack(udp *UdpConnection, track *webrtc.TrackRemote) error {
	for {
		select {
		case <-f.quit:
			slog.Info("forwarder closed", "forwarderId", f.id)
			return nil
		default:
			var err error
			b := make([]byte, 1500)
			rtpPacket := &rtp.Packet{}
			for {
				// Read
				n, _, readErr := track.Read(b)
				if readErr != nil {
					panic(readErr)
				}

				// Unmarshal the packet and update the PayloadType
				if err = rtpPacket.Unmarshal(b[:n]); err != nil {
					return fmt.Errorf("unmarshaling pkg: %w", err)
				}
				rtpPacket.PayloadType = udp.payloadType

				// Marshal into original buffer with updated PayloadType
				if n, err = rtpPacket.MarshalTo(b); err != nil {
					return fmt.Errorf("marshaling pkg: %w", err)
				}

				// Write
				if _, writeErr := udp.conn.Write(b[:n]); writeErr != nil {
					// For this particular example, third party applications usually timeout after a short
					// amount of time during which the user doesn't have enough time to provide the answer
					// to the browser.
					// That's why, for this particular example, the user first needs to provide the answer
					// to the browser then open the third party application. Therefore we must not kill
					// the forward on "connection refused" errors
					var opError *net.OpError
					if errors.As(writeErr, &opError) && opError.Err.Error() == "write: connection refused" {
						continue
					}
					return fmt.Errorf("writing pkg: %w", writeErr)
				}
			}
		}
	}
	return nil
}
