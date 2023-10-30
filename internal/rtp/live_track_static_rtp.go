package rtp

import (
	"errors"
	"strings"
	"sync"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

type trackBinding struct {
	id          string
	ssrc        webrtc.SSRC
	payloadType webrtc.PayloadType
	writeStream webrtc.TrackLocalWriter
}

// LiveTrackStaticRTP  is a TrackLocal that has a pre-set codec and accepts RTP Packets.
type LiveTrackStaticRTP struct {
	mu           sync.RWMutex
	bindings     []trackBinding
	codec        webrtc.RTPCodecCapability
	id, streamID string
}

// NewLiveTrackStaticRTP returns a LiveTrackStaticRTP.
func NewLiveTrackStaticRTP(c webrtc.RTPCodecCapability, id, streamID string) (*LiveTrackStaticRTP, error) {
	return &LiveTrackStaticRTP{
		codec:    c,
		bindings: []trackBinding{},
		id:       id,
		streamID: streamID,
	}, nil
}

func (s *LiveTrackStaticRTP) Bind(t webrtc.TrackLocalContext) (webrtc.RTPCodecParameters, error) {
	return webrtc.RTPCodecParameters{}, nil
}

func (s *LiveTrackStaticRTP) Unbind(t webrtc.TrackLocalContext) error {
	return nil
}

func (s *LiveTrackStaticRTP) BindTrack(t trackBinding) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.bindings = append(s.bindings, t)

	return nil
}

func (s *LiveTrackStaticRTP) UnbindTrack(t trackBinding) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.bindings {
		if s.bindings[i].id == t.id {
			s.bindings[i] = s.bindings[len(s.bindings)-1]
			s.bindings = s.bindings[:len(s.bindings)-1]
			return nil
		}
	}

	return errors.New("unbind error")
}

func (s *LiveTrackStaticRTP) ID() string { return s.id }

// StreamID is the group this track belongs too. This must be unique
func (s *LiveTrackStaticRTP) StreamID() string { return s.streamID }

// Kind controls if this TrackLocal is audio or video
func (s *LiveTrackStaticRTP) Kind() webrtc.RTPCodecType {
	switch {
	case strings.HasPrefix(s.codec.MimeType, "audio/"):
		return webrtc.RTPCodecTypeAudio
	case strings.HasPrefix(s.codec.MimeType, "video/"):
		return webrtc.RTPCodecTypeVideo
	default:
		return webrtc.RTPCodecType(0)
	}
}

// Codec gets the Codec of the track
func (s *LiveTrackStaticRTP) Codec() webrtc.RTPCodecCapability {
	return s.codec
}

// packetPool is a pool of packets used by WriteRTP and Write below
// nolint:gochecknoglobals
var rtpPacketPool = sync.Pool{
	New: func() interface{} {
		return &rtp.Packet{}
	},
}

// WriteRTP writes a RTP Packet to the LiveTrackStaticRTP
// If one PeerConnection fails the packets will still be sent to
// all PeerConnections. The error message will contain the ID of the failed
// PeerConnections so you can remove them
func (s *LiveTrackStaticRTP) WriteRTP(p *rtp.Packet) error {
	ipacket := rtpPacketPool.Get()
	packet := ipacket.(*rtp.Packet)
	defer func() {
		*packet = rtp.Packet{}
		rtpPacketPool.Put(ipacket)
	}()
	*packet = *p
	return s.writeRTP(packet)
}

// writeRTP is like WriteRTP, except that it may modify the packet p
func (s *LiveTrackStaticRTP) writeRTP(p *rtp.Packet) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	writeErrs := []error{}

	for _, b := range s.bindings {
		//p.Header.SSRC = uint32(b.ssrc)
		p.Header.PayloadType = uint8(b.payloadType)
		if _, err := b.writeStream.WriteRTP(&p.Header, p.Payload); err != nil {
			writeErrs = append(writeErrs, err)
		}
	}

	return flattenErrs(writeErrs)
}

// Write writes a RTP Packet as a buffer to the LiveTrackStaticRTP
// If one PeerConnection fails the packets will still be sent to
// all PeerConnections. The error message will contain the ID of the failed
// PeerConnections so you can remove them
func (s *LiveTrackStaticRTP) Write(b []byte) (n int, err error) {
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
