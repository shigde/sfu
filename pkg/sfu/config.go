package sfu

import (
	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/logging"
	"github.com/shigde/sfu/pkg/metric"
	"github.com/shigde/sfu/pkg/rtp"
	"github.com/shigde/sfu/pkg/storage"
)

type Config struct {
	*auth.AuthConfig       `mapstructure:"auth"`
	*storage.StorageConfig `mapstructure:"store"`
	*logging.LogConfig     `mapstructure:"log"`
	*metric.MetricConfig   `mapstructure:"metric"`
	*rtp.RtpConfig         `mapstructure:"rtp"`
}
