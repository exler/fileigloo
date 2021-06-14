package cmd

import (
	"log"
	"time"

	"github.com/getsentry/sentry-go"
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
	viper.SetDefault("storage", "local")

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

	err := sentry.Init(sentry.ClientOptions{
		TracesSampleRate: 0.2,
		TracesSampler: sentry.TracesSamplerFunc(func(ctx sentry.SamplingContext) sentry.Sampled {
			return sentry.SampledTrue
		}),
	})
	if err != nil {
		log.Fatalf("Sentry initialization error: %s", err)
	}
	// Flush buffered events before the program terminates
	defer sentry.Flush(2 * time.Second)

	Cmd.AddCommand(versionCmd, serverCmd)
}

func Execute() error {
	return Cmd.Execute()
}
