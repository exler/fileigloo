package cmd

import (
	"log"
	"os"

	"github.com/exler/fileigloo/server"
	"github.com/exler/fileigloo/storage"
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
				server.Port(viper.GetInt("PORT")),
				server.MaxUploadSize(viper.GetInt64("MAX_UPLOAD_SIZE")),
				server.RateLimit(viper.GetInt("RATE_LIMIT")),
			}

			switch storageProvider := viper.GetString("STORAGE"); storageProvider {
			case "local":
				if udir := viper.GetString("UPLOAD_DIRECTORY"); udir == "" {
					log.Println("Upload directory must be set for local storage!")
					os.Exit(0)
				} else if storage, err := storage.NewLocalStorage(udir); err != nil {
					log.Println(err)
					os.Exit(1)
				} else {
					serverOptions = append(serverOptions, server.UseStorage(storage))
				}
			case "s3":
				bucket := viper.GetString("S3_BUCKET")
				region := viper.GetString("S3_REGION")
				accessKey := viper.GetString("AWS_ACCESS_KEY")
				secretKey := viper.GetString("AWS_SECRET_KEY")
				sessionToken := viper.GetString("AWS_SESSION_TOKEN")
				endpointUrl := viper.GetString("AWS_ENDPOINT_URL")

				if storage, err := storage.NewS3Storage(accessKey, secretKey, sessionToken, endpointUrl, region, bucket); err != nil {
					log.Println(err)
					os.Exit(1)
				} else {
					serverOptions = append(serverOptions, server.UseStorage(storage))
				}
			default:
				log.Println("Incorrect or no storage type chosen!")
				os.Exit(0)
			}

			srv := server.New(serverOptions...)
			srv.Run()
		},
	}
)

func init() {
	serverCmd.Flags().Int("port", 8000, "Port to run the server on")
	serverCmd.Flags().Bool("https-only", false, "Automatically make all URLs with HTTPS schema")
	viper.BindPFlag("port", serverCmd.Flags().Lookup("port"))             //#nosec
	viper.BindPFlag("https-only", serverCmd.Flags().Lookup("https-only")) //#nosec
}
