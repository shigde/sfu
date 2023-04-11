package config

import (
	"fmt"
	"os"

	"github.com/shigde/sfu/pkg/auth"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	*auth.AuthConfig `mapstructure:"auth"`
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

	return config, nil
}
