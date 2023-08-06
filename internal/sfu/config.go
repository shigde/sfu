package sfu

import (
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/logging"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/storage"
)

type Config struct {
	*auth.SecurityConfig   `mapstructure:"security"`
	*storage.StorageConfig `mapstructure:"store"`
	*logging.LogConfig     `mapstructure:"log"`
	*metric.MetricConfig   `mapstructure:"metric"`
	*rtp.RtpConfig         `mapstructure:"rtp"`
}
