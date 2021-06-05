package cmd

import (
	"log"

	"github.com/exler/fileigloo/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	serverCmd = &cobra.Command{
		Use:   "runserver",
		Short: "Run web server",
		Long:  "Run web server allowing to upload files and pastes via API or browser",
		Run: func(cmd *cobra.Command, args []string) {
			serverOptions := []server.OptionFn{
				server.Port(viper.GetInt("Port")),
				server.MaxUploadSize(viper.GetInt64("max_upload_size")),
				server.RateLimit(viper.GetInt("rate_limit")),
				server.Purge(viper.GetInt("purge_older"), viper.GetInt("purge_interval")),
			}

			switch storageProvider := viper.GetString("storage"); storageProvider {
			case "local":
				if udir := viper.GetString("upload_directory"); udir == "" {
					log.Fatalln("Upload directory must be set for local storage!")
				} else if storage, err := server.NewLocalStorage(udir); err != nil {
					log.Fatalln(err)
				} else {
					serverOptions = append(serverOptions, server.UseStorage(storage))
				}
			}

			srv := server.New(serverOptions...)
			if err := srv.Run(); err != nil {
				log.Fatalln(err)
			}
		},
	}
)

func init() {
	serverCmd.Flags().Int("port", 8000, "Port to run the server on")
	viper.BindPFlag("port", serverCmd.Flags().Lookup("port"))
}
