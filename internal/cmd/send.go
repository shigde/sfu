package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	video string
	audio string

	sendCmd = &cobra.Command{
		Use:   "send",
		Short: "Send media to a Shig Lobby",
		Long:  ``,
		Run:   send,
	}
)

func send(ccmd *cobra.Command, args []string) {
	if len(args) > 0 {

	} else {
		fmt.Fprintln(os.Stderr, "No video and audio is specified. Please specify a valid Video and Audio file")
		return
	}
}

func init() {
	includeAddFlags(sendCmd)
}

func includeAddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&audio, "audio", "input.ogg", "audio.")
	cmd.PersistentFlags().StringVar(&video, "video", "input.ivf", "video")
	viper.BindPFlag("audio", cmd.PersistentFlags().Lookup("audio"))
	viper.BindPFlag("video", cmd.PersistentFlags().Lookup("video"))
}
