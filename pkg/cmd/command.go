package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	shigCtl = &cobra.Command{
		Use:           "shigCtl",
		Short:         "shigCtl – command-line tool to interact with shig",
		Long:          ``,
		Version:       "0.0.1",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func Execute() error {
	return shigCtl.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	shigCtl.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.shigCtl.yaml)")
	_ = viper.BindPFlag("config", shigCtl.PersistentFlags().Lookup("config"))

	shigCtl.AddCommand(send)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigFile(".shigCtl")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Config file used for shigCtl: ", viper.ConfigFileUsed())
	}
}
