package rtp

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

type UdpForwarder struct {
	mu           sync.RWMutex
	id           uuid.UUID
	audio, video *UdpConnection
	quit         chan struct{}
}

func NewUdpForwarder(id uuid.UUID, quit chan struct{}) (*UdpForwarder, error) {
	f := &UdpForwarder{
		mu:    sync.RWMutex{},
		id:    id,
		audio: &UdpConnection{port: 4000, payloadType: 111},
		video: &UdpConnection{port: 4002, payloadType: 96},
		quit:  quit,
	}

	go f.Run()
	return f, nil
}

func (f *UdpForwarder) Run() {
	f.log("run")
	defer func() {
		f.log("stop")
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

	f.log("running")
	select {
	case <-f.quit:
	}
	f.log("quit")
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

//func (f *UdpForwarder) AddTrack(trackInfo *TrackInfo) {
//
//	if trackInfo.RemoteTrack.Kind() == webrtc.RTPCodecTypeAudio {
//		go func(track *TrackInfo) {
//			f.log("writing audio")
//			if err := f.writeTrack(f.audio, track.RemoteTrack); err != nil {
//				slog.Error("writing stream audio", "err", err, "forwarderId", f.id.String())
//			}
//			f.log("stop writing stream audio")
//		}(trackInfo)
//	}
//
//	if trackInfo.RemoteTrack.Kind() == webrtc.RTPCodecTypeVideo {
//		go func(track *TrackInfo) {
//			f.log("writing video")
//			if err := f.writeTrack(f.video, track.RemoteTrack); err != nil {
//				slog.Error("writing stream video", "err", err, "forwarderId", f.id.String())
//			}
//			f.log("stop writing stream video")
//		}(trackInfo)
//	}
//}

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

	f.log(fmt.Sprintf("connected to port %d", udp.port))
	return err
}

func (f *UdpForwarder) close() error {
	if f.audio.conn != nil {
		f.log("close audio connection")
		if err := f.audio.conn.Close(); err != nil {
			return fmt.Errorf("closing audio udp port: %w", err)
		}
	}
	if f.video.conn != nil {
		f.log("close video connection")
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
			f.log("stop writing because quit")
			return nil
		default:
			var err error
			b := make([]byte, 1500)
			rtpPacket := &rtp.Packet{}
			for {
				// Read
				n, _, readErr := track.Read(b)
				if readErr != nil {
					f.log(fmt.Sprintf("can not anymore read %s track %s  stream %s", track.Kind(), track.ID(), track.StreamID()))
					return nil
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

func (f *UdpForwarder) GetConnData() UdpShare {
	return UdpShare{
		Audio: UdpShareInfo{Port: f.audio.port, PayloadType: f.audio.payloadType},
		Video: UdpShareInfo{Port: f.video.port, PayloadType: f.video.payloadType},
	}
}

func (f *UdpForwarder) log(msg string) {
	slog.Debug(msg, "forwarderId", f.id, "obj", "udpForwarder")
}

func (f *UdpForwarder) AddTrack(track *LiveTrackStaticRTP) {
	f.mu.Lock()
	defer f.mu.Unlock()

	binding := trackBinding{
		id: uuid.NewString(),
	}

	if track.Kind() == webrtc.RTPCodecTypeAudio {
		binding.ssrc = webrtc.SSRC(3450704251)
		binding.payloadType = webrtc.PayloadType(f.audio.payloadType)
		binding.writeStream = newUdpWriter(uuid.NewString(), f.audio, f.quit)
	}
	if track.Kind() == webrtc.RTPCodecTypeVideo {
		binding.ssrc = webrtc.SSRC(3450704222)
		binding.payloadType = webrtc.PayloadType(f.video.payloadType)
		binding.writeStream = newUdpWriter(uuid.NewString(), f.video, f.quit)
	}

	if err := track.BindTrack(binding); err != nil {
		slog.Error("binding track", "err", err)
	}
}
