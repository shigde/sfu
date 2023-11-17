package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	config  *Config

	shigClt = &cobra.Command{
		Use:           "shigClt",
		Short:         "shigClt – command-line tool to interact with shig",
		Long:          ``,
		Version:       "0.0.1",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func Execute() error {
	return shigClt.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	shigClt.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.shigClt.toml)")
	_ = viper.BindPFlag("config", shigClt.PersistentFlags().Lookup("config"))

	shigClt.AddCommand(sendCmd)
}

func initConfig() {
	config = &Config{}
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("toml")
		viper.SetConfigFile(".shigClt")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Config file used for shigClt: ", viper.ConfigFileUsed())
	}
	if err := viper.GetViper().Unmarshal(config); err != nil {
		fmt.Println("Config file used for shigClt: ", viper.ConfigFileUsed())
	}
}
