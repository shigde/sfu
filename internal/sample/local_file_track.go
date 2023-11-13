package sample

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/pion/webrtc/v3"
)

var ErrCannotDetermineMime = errors.New("cannot determine mime")

func NewLocalFileReaderTrack(file string, options ...ReaderOption) (*LocalTrack, error) {
	fp, mime, err := readFile(file)
	if err != nil {
		return nil, err
	}

	track, err := NewLocalReaderTrack(fp, mime, options...)
	if err != nil {
		_ = fp.Close()
		return nil, err
	}
	return track, nil
}

func NewLocalFileLooperTrack(file string, options ...ReaderOption) (*LocalTrack, error) {
	fp, mime, err := readFile(file)
	if err != nil {
		return nil, err
	}

	var looper Looper

	switch mime {
	case webrtc.MimeTypeH264:
	case webrtc.MimeTypeVP8:
	case webrtc.MimeTypeOpus:
	case webrtc.MimeTypeVP9:
	// allow
	default:
		return nil, ErrUnsupportedFileType
	}

	if spec.codec == h264Codec {
		looper, err := NewVideoLooperH264(f, spec)
		if err != nil {
			return nil, err
		}
		loopers = append(loopers, looper)
	} else if spec.codec == vp8Codec {
		looper, err := NewVideoLooperVP8(f, spec)
		if err != nil {
			return nil, err
		}
		loopers = append(loopers, looper)
	}

	track, err := NewLocalReaderTrack(fp, mime, options...)
	if err != nil {
		_ = fp.Close()
		return nil, err
	}
	return track, nil
}

func readFile(file string) (io.ReadCloser, string, error) {
	var err error
	if _, err = os.Stat(file); err != nil {
		return nil, "", err
	}

	// Open the file
	fp, err := os.Open(file)
	if err != nil {
		return nil, "", err
	}

	// Determine mime type from extension
	var mime string
	switch filepath.Ext(file) {
	case ".h264":
		mime = webrtc.MimeTypeH264
	case ".ivf":
		buf := make([]byte, 3)
		_, err = fp.ReadAt(buf, 8)
		if err != nil {
			return nil, "", err
		}
		switch string(buf) {
		case "VP8":
			mime = webrtc.MimeTypeVP8
		case "VP9":
			mime = webrtc.MimeTypeVP9
		default:
			_ = fp.Close()
			return nil, "", ErrCannotDetermineMime
		}
		_, _ = fp.Seek(0, 0)
	case ".ogg":
		mime = webrtc.MimeTypeOpus
	default:
		_ = fp.Close()
		return nil, "", ErrCannotDetermineMime
	}
	return fp, mime, nil
}
