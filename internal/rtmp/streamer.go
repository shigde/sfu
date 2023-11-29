package rtmp

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
)

type Streamer struct {
	quit chan struct{}
}

func NewStreamer(quit chan struct{}) *Streamer {
	return &Streamer{
		quit: quit,
	}
}

func (s *Streamer) StartFFmpeg(ctx context.Context, streamURL string) error {
	// Create a ffmpeg process that consumes MKV via stdin, and broadcasts out to Stream URL
	ffmpeg := exec.CommandContext(ctx, "ffmpeg", "-protocol_whitelist", "file,udp,rtp", "-i", "rtp-forwarder.sdp", "-c:v", "libx264", "-preset", "veryfast", "-b:v", "3000k", "-maxrate", "3000k", "-bufsize", "6000k", "-pix_fmt", "yuv420p", "-g", "50", "-c:a", "aac", "-b:a", "160k", "-ac", "2", "-ar", "44100", "-f", "flv", streamURL) //nolint
	if _, err := ffmpeg.StdinPipe(); err != nil {
		return fmt.Errorf("piping ffmpeg: %w", err)
	}
	ffmpegOut, _ := ffmpeg.StderrPipe()
	if err := ffmpeg.Start(); err != nil {
		return fmt.Errorf("starting ffmpeg: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(ffmpegOut)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
			//if ctx.Err() == context.Canceled {
			//	break
			//}
			select {
			case <-ctx.Done():
				return
			case <-s.quit:
				return
			default:
			}
		}
	}()
	return nil
}

func (s *Streamer) Stop() {
	select {
	case <-s.quit:
		return
	default:
		close(s.quit)
		<-s.quit
	}
}
