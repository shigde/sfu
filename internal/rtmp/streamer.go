package rtmp

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
)

func StartFFmpeg(ctx context.Context, streamURL string) {
	// Create a ffmpeg process that consumes MKV via stdin, and broadcasts out to Stream URL
	ffmpeg := exec.CommandContext(ctx, "ffmpeg", "-protocol_whitelist", "file,udp,rtp", "-i", "rtp-forwarder.sdp", "-c:v", "copy", "-c:a", "aac", "-f", "flv", "-flvflags", "no_duration_filesize", "-c:v", "libx264", streamURL) //nolint
	ffmpeg.StdinPipe()
	ffmpegOut, _ := ffmpeg.StderrPipe()
	if err := ffmpeg.Start(); err != nil {
		panic(err)
	}

	go func() {
		scanner := bufio.NewScanner(ffmpegOut)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
			if ctx.Err() == context.Canceled {
				break
			}
		}
	}()
}
