package rtp

//
//import (
//	"fmt"
//	"net"
//	"sync"
//
//	"github.com/google/uuid"
//	"github.com/pion/webrtc/v3"
//	"golang.org/x/exp/slog"
//)
//
//type UdpEndpoint struct {
//	mu           sync.RWMutex
//	id           uuid.UUID
//	streamId     string
//	audio, video *UdpConnection
//	quit         chan struct{}
//}
//
//func NewUdpEndpoint(id uuid.UUID, streamId string, quit chan struct{}) (*UdpEndpoint, error) {
//	ue := &UdpEndpoint{
//		mu:       sync.RWMutex{},
//		id:       id,
//		streamId: streamId,
//		audio:    &UdpConnection{port: 4000, payloadType: 111},
//		video:    &UdpConnection{port: 4002, payloadType: 96},
//		quit:     quit,
//	}
//
//	go ue.Run()
//	return ue, nil
//}
//
//func (ue *UdpEndpoint) Run() {
//	ue.mu.Lock()
//	ue.log("run")
//
//	defer func() {
//		ue.log("stop")
//		if err := ue.close(); err != nil {
//			slog.Error("udpEndpoint closing udp ports", "err", err, "udpEndpointID", ue.id)
//		}
//	}()
//
//	var err error
//	var laddr *net.UDPAddr
//	if laddr, err = net.ResolveUDPAddr("udp", "127.0.0.1:"); err != nil {
//		ue.mu.Unlock()
//		slog.Error("udpEndpoint: resolving udp address", "err", err, "udpEndpointID", ue.id.String())
//		return
//	}
//
//	if err = ue.connect(ue.audio, laddr); err != nil {
//		ue.mu.Unlock()
//		slog.Error("udpEndpoint: connecting audio", "err", err, "udpEndpointID", ue.id.String())
//		return
//	}
//	if err = ue.connect(ue.video, laddr); err != nil {
//		ue.mu.Unlock()
//		slog.Error("udpEndpoint: connecting video", "err", err, "udpEndpointID", ue.id.String())
//		return
//	}
//
//	ue.mu.Unlock()
//	ue.log("running")
//	select {
//	case <-ue.quit:
//	}
//	ue.log("quit")
//}
//
//func (ue *UdpEndpoint) connect(udp *UdpConnection, laddr *net.UDPAddr) error {
//	// Create remote addr
//	var raddr *net.UDPAddr
//	var err error
//	if raddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", udp.port)); err != nil {
//		return fmt.Errorf("resolving udp port: %w", err)
//	}
//	// Dial udp
//	if udp.conn, err = net.DialUDP("udp", laddr, raddr); err != nil {
//		return fmt.Errorf("dealing udp port: %w", err)
//	}
//
//	ue.log(fmt.Sprintf("connected to port %d", udp.port))
//	return err
//}
//
//func (ue *UdpEndpoint) close() error {
//	if ue.audio.conn != nil {
//		ue.log("close audio connection")
//		if err := ue.audio.conn.Close(); err != nil {
//			return fmt.Errorf("closing audio udp port: %w", err)
//		}
//	}
//	if ue.video.conn != nil {
//		ue.log("close video connection")
//		if err := ue.video.conn.Close(); err != nil {
//			return fmt.Errorf("closing video udp port: %w", err)
//		}
//	}
//	return nil
//}
//
//func (ue *UdpEndpoint) Stop() {
//	select {
//	case <-ue.quit:
//		return
//	default:
//		close(ue.quit)
//		<-ue.quit
//	}
//}
//
//func (ue *UdpEndpoint) log(msg string) {
//	slog.Debug(msg, "udpEndpointID", ue.id, "obj", "UdpEndpoint")
//}
//
//func (ue *UdpEndpoint) GetConnData() UdpShare {
//	return UdpShare{
//		Audio: UdpShareInfo{Port: ue.audio.port, PayloadType: ue.audio.payloadType},
//		Video: UdpShareInfo{Port: ue.video.port, PayloadType: ue.video.payloadType},
//	}
//}
//
//func (ue *UdpEndpoint) AddTrack(track webrtc.TrackLocal) {
//	ue.mu.Lock()
//	defer ue.mu.Unlock()
//
//	if r.hasSent() {
//		return err
//	}
//
//	writeStream := &interceptorToTrackLocalWriter{}
//	r.context = TrackLocalContext{
//		id:          r.id,
//		params:      r.api.mediaEngine.getRTPParametersByKind(r.track.Kind(), []RTPTransceiverDirection{RTPTransceiverDirectionSendonly}),
//		ssrc:        parameters.Encodings[0].SSRC,
//		writeStream: writeStream,
//	}
//
//	codec, err := r.track.Bind(r.context)
//	if err != nil {
//		return err
//	}
//	//r.context.params.Codecs = []RTPCodecParameters{codec}
//	//
//	//r.streamInfo = *createStreamInfo(r.id, parameters.Encodings[0].SSRC, codec.PayloadType, codec.RTPCodecCapability, parameters.HeaderExtensions)
//	//rtpInterceptor := r.api.interceptor.BindLocalStream(&r.streamInfo, interceptor.RTPWriterFunc(func(header *rtp.Header, payload []byte, attributes interceptor.Attributes) (int, error) {
//	//	return r.srtpStream.WriteRTP(header, payload)
//	//}))
//	//writeStream.interceptor.Store(rtpInterceptor)
//
//}
//
//
//func (f *UdpForwarder) buildContext(kind any) TrackLocalContext {
//	var ssrc webrtc.SSRC
//	params := webrtc.RTPParameters{
//		HeaderExtensions: make([]webrtc.RTPHeaderExtensionParameter, 0),
//		Codecs:           make([]webrtc.RTPCodecParameters, 0),
//	}
//	switch kind {
//	case "audio":
//		ssrc = webrtc.SSRC(3450704251)
//		params.Codecs = append(params.Codecs, f.getAudioCodecs())
//	case "video":
//		ssrc = webrtc.SSRC(3450704240)
//		params.Codecs = append(params.Codecs, f.getVideoCodecs())
//	}
//
//	return TrackLocalContext{
//		id:     uuid.NewString(),
//		params: params,
//		ssrc:   ssrc,
//		//writeStream: writeStream,
//	}
//}
//
//func (f *UdpForwarder) getVideoCodecs() webrtc.RTPCodecParameters {
//	return webrtc.RTPCodecParameters{
//		RTPCodecCapability: webrtc.RTPCodecCapability{
//			MimeType:     webrtc.MimeTypeVP8,
//			ClockRate:    90000,
//			Channels:     0,
//			SDPFmtpLine:  "",
//			RTCPFeedback: nil,
//		},
//	}
//}
//func (f *UdpForwarder) getAudioCodecs() webrtc.RTPCodecParameters {
//	return webrtc.RTPCodecParameters{
//		RTPCodecCapability: webrtc.RTPCodecCapability{
//			MimeType:     webrtc.MimeTypeOpus,
//			ClockRate:    48000,
//			Channels:     0,
//			SDPFmtpLine:  "",
//			RTCPFeedback: nil,
//		},
//	}
//}
//
//type TrackLocalContext struct {
//	id          string
//	params      webrtc.RTPParameters
//	ssrc        webrtc.SSRC
//	writeStream webrtc.TrackLocalWriter
//}
//
//// CodecParameters returns the negotiated RTPCodecParameters. These are the codecs supported by both
//// PeerConnections and the SSRC/PayloadTypes
//func (t *TrackLocalContext) CodecParameters() []webrtc.RTPCodecParameters {
//	return t.params.Codecs
//}
//
//// HeaderExtensions returns the negotiated RTPHeaderExtensionParameters. These are the header extensions supported by
//// both PeerConnections and the SSRC/PayloadTypes
//func (t *TrackLocalContext) HeaderExtensions() []webrtc.RTPHeaderExtensionParameter {
//	return t.params.HeaderExtensions
//}
//
//// SSRC requires the negotiated SSRC of this track
//// This track may have multiple if RTX is enabled
//func (t *TrackLocalContext) SSRC() webrtc.SSRC {
//	return t.ssrc
//}
//
//// WriteStream returns the WriteStream for this TrackLocal. The implementer writes the outbound
//// media packets to it
//func (t *TrackLocalContext) WriteStream() webrtc.TrackLocalWriter {
//	return t.writeStream
//}
//
//// ID is a unique identifier that is used for both Bind/Unbind
//func (t *TrackLocalContext) ID() string {
//	return t.id
//}
