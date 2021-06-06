package cmd

import (
	"log"

	colors "github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &cobra.Command{
	Use:   "fileigloo",
	Short: "Small and simple temporary file sharing & pastebin",
	Long: `Small and simple temporary file sharing & pastebin. 
Source code available at github.com/exler/fileigloo`,
}

func init() {
	log.SetPrefix(colors.Blue("[fileigloo] ").String())

	viper.AutomaticEnv()
	viper.AddConfigPath("config/")
	viper.SetConfigName("fileigloo")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalln("Configuration file not found")
		} else {
			log.Fatalln("Unable to load configuration file")
		}
	}

	Cmd.AddCommand(versionCmd, serverCmd)
}

func Execute() error {
	return Cmd.Execute()
}
