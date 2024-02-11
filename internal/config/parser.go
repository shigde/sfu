package config

import (
	"fmt"
	"os"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/sfu"
	"github.com/shigde/sfu/internal/telemetry"
	"github.com/spf13/viper"
)

func ParseConfig(file string, env *sfu.Environment) (*sfu.Config, error) {
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

	if config.StorageConfig.Name != "sqlite3" {
		return nil, fmt.Errorf("store.name currently supportes only Sqlite3")
	}

	if len(config.StorageConfig.DataSource) == 0 {
		return nil, fmt.Errorf("store.dataSource should not be empty")
	}

	if len(config.LogConfig.Logfile) == 0 {
		return nil, fmt.Errorf("log.logfile should not be empty")
	}

	if err := metric.ValidateMetricConfig(config.MetricConfig); err != nil {
		return nil, err
	}

	if err := auth.ValidateSecurityConfig(config.SecurityConfig); err != nil {
		return nil, err
	}

	if err := rtp.ValidateRtpConfig(config.RtpConfig); err != nil {
		return nil, err
	}

	if err := instance.ValidateFederationConfig(config.FederationConfig, env.FederationEnv); err != nil {
		return nil, err
	}

	if err := sfu.ValidateServerConfig(config.ServerConfig); err != nil {
		return nil, err
	}

	if err := telemetry.ValidateTelemetryConfig(config.TelemetryConfig); err != nil {
		return nil, err
	}

	return config, nil
}
