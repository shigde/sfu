package sfu

import (
	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/logging"
	"github.com/shigde/sfu/pkg/metric"
)

type Config struct {
	*auth.AuthConfig     `mapstructure:"auth"`
	*logging.LogConfig   `mapstructure:"log"`
	*metric.MetricConfig `mapstructure:"metric"`
}
