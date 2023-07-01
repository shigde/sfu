package sfu

import (
	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/logging"
	"github.com/shigde/sfu/pkg/metric"
	"github.com/shigde/sfu/pkg/store"
)

type Config struct {
	*auth.AuthConfig     `mapstructure:"auth"`
	*store.StorageConfig `mapstructure:"store"`
	*logging.LogConfig   `mapstructure:"log"`
	*metric.MetricConfig `mapstructure:"metric"`
}
