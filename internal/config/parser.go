package config

import (
	"fmt"
	"os"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/auth/session"
	"github.com/shigde/sfu/internal/mail"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/storage"
	"github.com/shigde/sfu/internal/telemetry"
	"github.com/spf13/viper"
)

func ParseConfig(file string, env *Environment) (*SFU, error) {
	config := &SFU{}

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

	if err := storage.ValidateStorageConfig(config.StorageConfig); err != nil {
		return nil, err
	}

	if len(config.LogConfig.Logfile) == 0 {
		return nil, fmt.Errorf("log.logfile should not be empty")
	}

	if err := metric.ValidateMetricConfig(config.MetricConfig); err != nil {
		return nil, err
	}

	if err := session.ValidateSecurityConfig(config.SecurityConfig); err != nil {
		return nil, err
	}

	if err := rtp.ValidateRtpConfig(config.RtpConfig); err != nil {
		return nil, err
	}

	if err := instance.ValidateFederationConfig(config.FederationConfig, &env.FederationEnv); err != nil {
		return nil, err
	}

	if err := ValidateServerConfig(config.ServerConfig, &env.ServerEnv); err != nil {
		return nil, err
	}

	if err := telemetry.ValidateTelemetryConfig(config.TelemetryConfig); err != nil {
		return nil, err
	}

	if err := mail.ValidateEmailConfig(config.MailConfig); err != nil {
		return nil, err
	}

	return config, nil
}
