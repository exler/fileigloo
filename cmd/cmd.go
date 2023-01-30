package cmd

import (
	"log"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &cobra.Command{
	Use:   "fileigloo",
	Short: "Small and simple online file sharing & pastebin",
	Long: `Small and simple online file sharing & pastebin. 
Source code available at github.com/exler/fileigloo`,
}

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("STORAGE", "local")
	viper.SetDefault("UPLOAD_DIRECTORY", "uploads/")
	viper.SetDefault("RATE_LIMIT", 2)

	if sentry_dsn := viper.GetString("SENTRY_DSN"); sentry_dsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: sentry_dsn,
		})

		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
	}

	Cmd.AddCommand(versionCmd, serverCmd)
}

func Execute() error {
	return Cmd.Execute()
}
