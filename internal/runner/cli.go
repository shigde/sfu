package runner

import (
	"flag"
	"fmt"
	"os"
)

const (
	audioFileName = "output.ogg"
	videoFileName = "output.ivf"
	configToml    = "config.toml"
)

type Cli struct {
	AudioFileName string
	VideoFileName string
	Config        string
}

func NewCli() *Cli {
	return &Cli{}
}

func (cli *Cli) Parse() bool {
	var (
		conf  string
		audio string
		video string
	)

	flag.StringVar(&conf, "c", configToml, "Configuration toml file")
	flag.StringVar(&audio, "a", audioFileName, "Audio File")
	flag.StringVar(&video, "v", videoFileName, "Video File")
	help := flag.Bool("h", false, "help info")
	flag.Parse()

	if *help {
		return false
	}

	cli.Config = conf
	cli.AudioFileName = audio
	cli.VideoFileName = video

	return true
}

func (cli *Cli) ShowHelp() {
	fmt.Printf("Usage:%s [options...] <method>\n", os.Args[0])
	fmt.Println("Options: ---------")
	fmt.Println("          -h  help info")
	fmt.Printf("          -c  Configuration toml file (default: %s)\n", configToml)
	fmt.Printf("          -a  Audio File (default: %s)\n", audioFileName)
	fmt.Printf("          -v  Video File (default: %s)\n", videoFileName)
}
