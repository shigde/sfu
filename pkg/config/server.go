package config

import (
	"fmt"
	"os"

	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/logging"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	*auth.AuthConfig   `mapstructure:"auth"`
	*logging.LogConfig `mapstructure:"log"`
}

func ParseConfig(file string) (*ServerConfig, error) {
	config := &ServerConfig{}

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

	return config, nil
}
