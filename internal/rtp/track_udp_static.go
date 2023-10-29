package rtp

import (
	"sync"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

type TrackUdpStaticRTP struct {
	mu sync.RWMutex
	//bindings     []trackBinding
	codec        webrtc.RTPCodecCapability
	id, streamID string
}

// NewTrackUdpStaticRTP returns a TrackUdpStaticRTP.
func NewTrackUdpStaticRTP(c webrtc.RTPCodecCapability, id, streamID string) (*TrackUdpStaticRTP, error) {
	return &TrackUdpStaticRTP{
		codec:    c,
		id:       id,
		streamID: streamID,
	}, nil
}

// packetPool is a pool of packets used by WriteRTP and Write below
// nolint:gochecknoglobals
var rtpPacketPool = sync.Pool{
	New: func() interface{} {
		return &rtp.Packet{}
	},
}

func (s *TrackUdpStaticRTP) WriteRTP(p *rtp.Packet) error {
	ipacket := rtpPacketPool.Get()
	packet := ipacket.(*rtp.Packet)
	defer func() {
		*packet = rtp.Packet{}
		rtpPacketPool.Put(ipacket)
	}()
	*packet = *p
	return s.writeRTP(packet)
}

func (s *TrackUdpStaticRTP) Write(b []byte) (n int, err error) {
	ipacket := rtpPacketPool.Get()
	packet := ipacket.(*rtp.Packet)
	defer func() {
		*packet = rtp.Packet{}
		rtpPacketPool.Put(ipacket)
	}()

	if err = packet.Unmarshal(b); err != nil {
		return 0, err
	}

	return len(b), s.writeRTP(packet)
}

func (s *TrackUdpStaticRTP) writeRTP(p *rtp.Packet) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// writeErrs := []error{}

	//for _, b := range s.bindings {
	//	p.Header.SSRC = uint32(b.ssrc)
	//	p.Header.PayloadType = uint8(b.payloadType)
	//	if _, err := b.writeStream.WriteRTP(&p.Header, p.Payload); err != nil {
	//		writeErrs = append(writeErrs, err)
	//	}
	//}
	//
	//return util.FlattenErrs(writeErrs)
	return nil
}
