package sample

import (
	"bytes"
	"io"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"
	"github.com/shigde/sfu/internal/rtp"
	"golang.org/x/exp/slog"
)

type VideoLooperH264 struct {
	buffer        []byte
	frameDuration time.Duration
	spec          *videoSpec
	reader        *h264reader.H264Reader
}

func NewLocalLooperH264Track(input io.ReadCloser, mime string, spec *videoSpec, trackOpts LocalTrackOptions) (*LocalTrack, error) {
	looper, err := NewVideoLooperH264(input, spec)
	if err != nil {
		return nil, err
	}

	// Create sample track & bind handler
	track, err := NewLocalTrack(webrtc.RTPCodecCapability{MimeType: mime}, trackOpts)
	if err != nil {
		return nil, err
	}
	track.OnBind(func() {
		if err := track.StartWrite(looper, nil); err != nil {
			slog.Error("Could not start writing", "err", err)
		}
	})

	return track, nil
}

func NewVideoLooperH264(input io.Reader, spec *videoSpec) (*VideoLooperH264, error) {
	l := &VideoLooperH264{
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

func (l *VideoLooperH264) Codec() webrtc.RTPCodecCapability {
	return webrtc.RTPCodecCapability{
		MimeType:    "video/h264",
		ClockRate:   90000,
		Channels:    0,
		SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
		RTCPFeedback: []webrtc.RTCPFeedback{
			{Type: webrtc.TypeRTCPFBNACK},
			{Type: webrtc.TypeRTCPFBNACK, Parameter: "pli"},
		},
	}
}

func (l *VideoLooperH264) NextSample() (media.Sample, error) {
	return l.nextSample(true)
}

func (l *VideoLooperH264) ToLayer(quality rtp.VideoQuality) *rtp.SimulcastLayer {
	return l.spec.ToLayer(quality)
}

func (l *VideoLooperH264) nextSample(rewindEOF bool) (media.Sample, error) {
	sample := media.Sample{}
	if l.reader == nil {
		var err error
		l.reader, err = h264reader.NewReader(bytes.NewReader(l.buffer))
		if err != nil {
			return sample, err
		}
	}
	nal, err := l.reader.NextNAL()
	if err == io.EOF && rewindEOF {
		l.reader = nil
		return l.nextSample(false)
	}
	if err != nil {
		return sample, err
	}

	isFrame := false
	switch nal.UnitType {
	case h264reader.NalUnitTypeCodedSliceDataPartitionA,
		h264reader.NalUnitTypeCodedSliceDataPartitionB,
		h264reader.NalUnitTypeCodedSliceDataPartitionC,
		h264reader.NalUnitTypeCodedSliceIdr,
		h264reader.NalUnitTypeCodedSliceNonIdr:
		isFrame = true
	}

	sample.Data = nal.Data
	if isFrame {
		// return it without duration
		sample.Duration = l.frameDuration
	}
	return sample, nil
}

func (l *VideoLooperH264) OnBind() error {
	return nil
}

func (l *VideoLooperH264) OnUnbind() error {
	return nil
}

func (l *VideoLooperH264) Close() error {
	return nil
}
