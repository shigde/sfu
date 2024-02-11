package config

import (
	"github.com/shigde/sfu/internal/sfu"
	"github.com/spf13/viper"
)

func ParseEnv() *sfu.Environment {
	env := &sfu.Environment{}
	viper.SetEnvPrefix("shigde_instance") // will be uppercased automatically
	viper.BindEnv("domain")
	viper.BindEnv("register_token")

	env.FederationEnv.Domain = viper.GetString("domain")
	env.FederationEnv.RegisterToken = viper.GetString("register_token")
	return env
}
