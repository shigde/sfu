package sample

import (
	"bytes"
	"io"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggreader"
	"golang.org/x/exp/slog"
)

type AudioLooperOpus struct {
	buffer      []byte
	reader      *oggreader.OggReader
	lastGranule uint64
}

func NewLocalLooperOpusTrack(input io.ReadCloser, mime string, trackOpts LocalTrackOptions) (*LocalTrack, error) {
	looper, err := NewAudioLooperOpus(input)
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

func NewAudioLooperOpus(input io.Reader) (*AudioLooperOpus, error) {
	l := &AudioLooperOpus{}

	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, input); err != nil {
		return nil, err
	}
	l.buffer = buf.Bytes()

	return l, nil
}

func (l *AudioLooperOpus) Codec() webrtc.RTPCodecCapability {
	return webrtc.RTPCodecCapability{
		MimeType: "audio/opus",
	}
}

func (l *AudioLooperOpus) NextSample() (media.Sample, error) {
	return l.nextSample(true)
}

func (l *AudioLooperOpus) nextSample(rewindEOF bool) (media.Sample, error) {
	sample := media.Sample{}
	if l.reader == nil {
		var err error
		l.lastGranule = 0
		l.reader, _, err = oggreader.NewWith(bytes.NewReader(l.buffer))
		if err != nil {
			return sample, err
		}
	}

	pageData, pageHeader, err := l.reader.ParseNextPage()
	if err == io.EOF && rewindEOF {
		l.reader = nil
		return l.nextSample(false)
	}
	if err != nil {
		return sample, err
	}
	sampleCount := float64(pageHeader.GranulePosition - l.lastGranule)
	l.lastGranule = pageHeader.GranulePosition

	sample.Data = pageData
	sample.Duration = time.Duration((sampleCount/48000)*1000) * time.Millisecond
	if sample.Duration == 0 {
		sample.Duration = defaultOpusFrameDuration
	}
	return sample, nil
}

func (l *AudioLooperOpus) OnBind() error {
	return nil
}

func (l *AudioLooperOpus) OnUnbind() error {
	return nil
}

func (l *AudioLooperOpus) Close() error {
	return nil
}
