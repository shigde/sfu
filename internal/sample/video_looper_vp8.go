package sample

import (
	"bytes"
	"io"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/ivfreader"
	"github.com/shigde/sfu/internal/rtp"
)

type VideoLooperVP8 struct {
	buffer        []byte
	frameDuration time.Duration
	spec          *videoSpec
	reader        *ivfreader.IVFReader
	ivfTimebase   float64
	lastTimestamp uint64
}

func NewVideoLooperVP8(input io.Reader, spec *videoSpec) (*VideoLooperVP8, error) {
	l := &VideoLooperVP8{
		spec:          spec,
		frameDuration: time.Second / time.Duration(spec.fps),
	}

	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, input); err != nil {
		return nil, err
	}
	l.buffer = buf.Bytes()

	return l, nil
}

func (l *VideoLooperVP8) Codec() webrtc.RTPCodecCapability {
	return webrtc.RTPCodecCapability{
		MimeType:  "video/vp8",
		ClockRate: 90000,
		RTCPFeedback: []webrtc.RTCPFeedback{
			{Type: webrtc.TypeRTCPFBNACK},
			{Type: webrtc.TypeRTCPFBNACK, Parameter: "pli"},
		},
	}
}

func (l *VideoLooperVP8) NextSample() (media.Sample, error) {
	return l.nextSample(true)
}

func (l *VideoLooperVP8) ToLayer(quality rtp.VideoQuality) *rtp.SimulcastLayer {
	return l.spec.ToLayer(quality)
}

func (l *VideoLooperVP8) nextSample(rewindEOF bool) (media.Sample, error) {
	sample := media.Sample{}
	if l.reader == nil {
		var err error
		var ivfheader *ivfreader.IVFFileHeader
		l.reader, ivfheader, err = ivfreader.NewWith(bytes.NewReader(l.buffer))
		if err != nil {
			return sample, err
		}
		l.ivfTimebase = float64(ivfheader.TimebaseNumerator) / float64(ivfheader.TimebaseDenominator)
	}

	frame, header, err := l.reader.ParseNextFrame()
	if err == io.EOF && rewindEOF {
		l.reader = nil
		return l.nextSample(false)
	}
	if err != nil {
		return sample, err
	}
	delta := header.Timestamp - l.lastTimestamp
	sample.Data = frame
	// this should be correct too, but we'll use the known frame-rates below
	sample.Duration = time.Duration(l.ivfTimebase*float64(delta)*1000) * time.Millisecond
	l.lastTimestamp = header.Timestamp
	sample.Duration = l.frameDuration
	return sample, nil
}

func (l *VideoLooperVP8) OnBind() error {
	return nil
}

func (l *VideoLooperVP8) OnUnbind() error {
	return nil
}

func (l *VideoLooperVP8) Close() error {
	return nil
}
