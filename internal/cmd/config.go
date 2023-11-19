package cmd

import "github.com/shigde/sfu/internal/rtp"

type ShigConfig struct {
	User          string `mapstructure:"user"`
	Pass          string `mapstructure:"pass"`
	RegisterToken string `mapstructure:"registerToken"`
}

type RtmpConfig struct {
	StreamKey string `mapstructure:"streamKey"`
	RtmpUrl   string `mapstructure:"rtmpUrl"`
}

type Config struct {
	*ShigConfig    `mapstructure:"shig"`
	*RtmpConfig    `mapstructure:"rtmp"`
	*rtp.RtpConfig `mapstructure:"rtp"`
}
