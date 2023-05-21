package config

import (
	"fmt"
	"os"

	"github.com/shigde/sfu/pkg/sfu"
	"github.com/spf13/viper"
)

func ParseConfig(file string) (*sfu.Config, error) {
	config := &sfu.Config{}

	if _, err := os.Stat(file); err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}

	viper.SetConfigFile(file)
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := viper.GetViper().Unmarshal(config); err != nil {
		return nil, fmt.Errorf("loading config file: %w", err)
	}

	if err := viper.GetViper().Unmarshal(config); err != nil {
		return nil, fmt.Errorf("loading config file: %w", err)
	}

	if len(config.AuthConfig.JWT.Key) < 1 {
		return nil, fmt.Errorf("auth.jwt.key should not be empty")
	}

	if len(config.LogConfig.Logfile) == 0 {
		return nil, fmt.Errorf("log.logfile should not be empty")
	}

	if len(config.MetricConfig.Prometheus.Endpoint) == 0 {
		return nil, fmt.Errorf("metric.prometheus.endpoint should not be empty")
	}

	return config, nil
}
