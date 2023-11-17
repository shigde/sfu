package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/shigde/sfu/internal/runner"
	"github.com/shigde/sfu/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	video             string
	audio             string
	sendLiveStreamUrl string
	isMainStream      bool

	sendCmd = &cobra.Command{
		Use:   "send",
		Short: "Send static media stream in a Shig Lobby",
		Long:  "Read video and audio media from a file and stream it to a Shig Lobby",
		Run:   send,
	}
)

func send(ccmd *cobra.Command, args []string) {
	params, err := NewShigParamsByUrl(sendLiveStreamUrl)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, fmt.Errorf("get url param: %w", err))
		return
	}

	lobby := client.NewLobbyApi(config.ShigConfig.User, params.URL)
	token, err := lobby.Login()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, fmt.Errorf("logging in: %w", err))
		return
	}

	sender := runner.NewStaticSender(config.RtpConfig, audio, video, params.Space, params.Stream, token.JWT)
	ctx, cancelCtx := context.WithCancel(ccmd.Context())
	defer cancelCtx()
	runChn := make(chan struct{})
	go func() {
		defer close(runChn)
		err := sender.Run(ctx)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, fmt.Errorf("logging in: %w", err))
		}
	}()

	select {
	case <-runChn:
		_, _ = fmt.Fprintln(os.Stderr, "finish sending")
	}
}

func init() {
	includeAddFlags(sendCmd)
}

func includeAddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&audio, "audio", "input.ogg", "Audio file as ogg format, send as stream in a lobby")
	cmd.PersistentFlags().StringVar(&video, "video", "input.ivf", "Video file as ivf format, send as stream in a lobby")
	cmd.PersistentFlags().StringVar(&sendLiveStreamUrl, "url", "", "Shig live stream rest endpoint url")
	cmd.PersistentFlags().BoolVar(&isMainStream, "main", false, "When set the stream will send as main stream")
	viper.BindPFlag("audio", cmd.PersistentFlags().Lookup("audio"))
	viper.BindPFlag("video", cmd.PersistentFlags().Lookup("video"))
	viper.BindPFlag("url", cmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("main", cmd.PersistentFlags().Lookup("main"))
}
