package cmd

import "github.com/shigde/sfu/internal/rtp"

type ShigConfig struct {
	User          string `mapstructure:"user"`
	Pass          string `mapstructure:"pass"`
	RegisterToken string `mapstructure:"registerToken"`
}

type Config struct {
	*ShigConfig    `mapstructure:"shig"`
	*rtp.RtpConfig `mapstructure:"rtp"`
}
