package sample

import (
	"errors"
	"io"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"
	"github.com/pion/webrtc/v3/pkg/media/ivfreader"
	"github.com/pion/webrtc/v3/pkg/media/oggreader"
	"golang.org/x/exp/slog"
)

const (
	// defaults to 30 fps
	defaultH264FrameDuration = 33 * time.Millisecond
	defaultOpusFrameDuration = 20 * time.Millisecond
)

var ErrUnsupportedFileType = errors.New("unsupported file type")

type Reader struct {
	// Configuration
	Mime            string
	FrameDuration   time.Duration
	OnWriteComplete func()
	AudioLevel      uint8
	trackOpts       []LocalTrackOptions

	// Allow various types of ingress
	reader io.ReadCloser

	// for vp8/vp9
	ivfReader     *ivfreader.IVFReader
	ivfTimebase   float64
	lastTimestamp uint64

	// for h264
	h264reader *h264reader.H264Reader

	// for ogg
	oggReader   *oggreader.OggReader
	lastGranule uint64
}

type ReaderOption func(*Reader)

func ReaderTrackWithMime(mime string) func(provider *Reader) {
	return func(provider *Reader) {
		provider.Mime = mime
	}
}

func ReaderTrackWithFrameDuration(duration time.Duration) func(provider *Reader) {
	return func(provider *Reader) {
		provider.FrameDuration = duration
	}
}

func ReaderTrackWithOnWriteComplete(f func()) func(provider *Reader) {
	return func(provider *Reader) {
		provider.OnWriteComplete = f
	}
}

func ReaderTrackWithRTCPHandler(f func(rtcp.Packet)) func(provider *Reader) {
	return func(provider *Reader) {
		provider.trackOpts = append(provider.trackOpts, WithRTCPHandler(f))
	}
}

// NewLocalReaderTrack uses io.ReadCloser interface to adapt to various ingress types
// - mime: has to be one of webrtc.MimeType... (e.g. webrtc.MimeTypeOpus)
func NewLocalReaderTrack(in io.ReadCloser, mime string, options ...ReaderOption) (*LocalTrack, error) {
	reader := &Reader{
		Mime:   mime,
		reader: in,
		// default audio level to be fairly loud
		AudioLevel: 15,
	}
	for _, opt := range options {
		opt(reader)
	}

	// check if mime type is supported
	switch reader.Mime {
	case webrtc.MimeTypeH264, webrtc.MimeTypeOpus, webrtc.MimeTypeVP8, webrtc.MimeTypeVP9:
	// allow
	default:
		return nil, ErrUnsupportedFileType
	}

	// Create sample track & bind handler
	track, err := NewLocalTrack(webrtc.RTPCodecCapability{MimeType: reader.Mime}, reader.trackOpts...)
	if err != nil {
		return nil, err
	}
	track.OnBind(func() {
		if err := track.StartWrite(reader, reader.OnWriteComplete); err != nil {
			slog.Error("Could not start writing", "err", err)
		}
	})

	return track, nil
}

func (p *Reader) OnBind() error {
	// If we are not closing on unbind, don't do anything on rebind
	if p.ivfReader != nil || p.h264reader != nil || p.oggReader != nil {
		return nil
	}

	var err error
	switch p.Mime {
	case webrtc.MimeTypeH264:
		p.h264reader, err = h264reader.NewReader(p.reader)
	case webrtc.MimeTypeVP8, webrtc.MimeTypeVP9:
		var ivfHeader *ivfreader.IVFFileHeader
		p.ivfReader, ivfHeader, err = ivfreader.NewWith(p.reader)
		if err == nil {
			p.ivfTimebase = float64(ivfHeader.TimebaseNumerator) / float64(ivfHeader.TimebaseDenominator)
		}
	case webrtc.MimeTypeOpus:
		p.oggReader, _, err = oggreader.NewWith(p.reader)
	default:
		err = ErrUnsupportedFileType
	}
	if err != nil {
		_ = p.reader.Close()
		return err
	}
	return nil
}

func (p *Reader) OnUnbind() error {
	return nil
}

func (p *Reader) Close() error {
	if p.reader != nil {
		return p.reader.Close()
	}
	return nil
}

func (p *Reader) CurrentAudioLevel() uint8 {
	return p.AudioLevel
}

func (p *Reader) NextSample() (media.Sample, error) {
	sample := media.Sample{}
	switch p.Mime {
	case webrtc.MimeTypeH264:
		nal, err := p.h264reader.NextNAL()
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
		if !isFrame {
			// return it without duration
			return sample, nil
		}
		sample.Duration = defaultH264FrameDuration
	case webrtc.MimeTypeVP8, webrtc.MimeTypeVP9:
		frame, header, err := p.ivfReader.ParseNextFrame()
		if err != nil {
			return sample, err
		}
		delta := header.Timestamp - p.lastTimestamp
		sample.Data = frame
		sample.Duration = time.Duration(p.ivfTimebase*float64(delta)*1000) * time.Millisecond
		p.lastTimestamp = header.Timestamp
	case webrtc.MimeTypeOpus:
		pageData, pageHeader, err := p.oggReader.ParseNextPage()
		if err != nil {
			return sample, err
		}
		sampleCount := float64(pageHeader.GranulePosition - p.lastGranule)
		p.lastGranule = pageHeader.GranulePosition

		sample.Data = pageData
		sample.Duration = time.Duration((sampleCount/48000)*1000) * time.Millisecond
		if sample.Duration == 0 {
			sample.Duration = defaultOpusFrameDuration
		}
	}

	if p.FrameDuration > 0 {
		sample.Duration = p.FrameDuration
	}
	return sample, nil
}
