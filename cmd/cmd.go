package cmd

import (
	"log"

	colors "github.com/logrusorgru/aurora"
	"github.com/rollbar/rollbar-go"
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
	viper.SetDefault("storage", "local")
	viper.SetDefault("upload_directory", "uploads/")
	viper.SetDefault("rate_limit", 2)
	viper.SetDefault("purge_older", 24)
	viper.SetDefault("purge_interval", 24)

	// Setup Rollbar logging
	rollbar.SetToken(viper.GetString("ROLLBAR_TOKEN"))
	rollbar.SetEnvironment(viper.GetString("ROLLBAR_ENVIRONMENT"))

	Cmd.AddCommand(versionCmd, serverCmd)
}

func Execute() error {
	return Cmd.Execute()
}
