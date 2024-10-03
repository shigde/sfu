package config

import (
	"github.com/spf13/viper"
)

func ParseEnv() *Environment {
	env := &Environment{}
	viper.SetEnvPrefix("shigde_instance") // will be uppercased automatically
	if err := viper.BindEnv("domain"); err != nil {
		return env
	}
	if err := viper.BindEnv("register_token"); err != nil {
		return env
	}
	if err := viper.BindEnv("port"); err != nil {
		return env
	}

	env.FederationEnv.Domain = viper.GetString("domain")
	env.FederationEnv.RegisterToken = viper.GetString("register_token")
	env.ServerEnv.Port = viper.GetInt("port")
	return env
}
