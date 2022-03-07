package server

import "github.com/shigde/sfu/pkg/auth"

type Config struct {
	Auth  auth.AuthConfig `mapstructure:"auth"`
}

func (c Config) GetAuth() *auth.AuthConfig {
	return &c.Auth
}
