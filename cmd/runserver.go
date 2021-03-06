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
				server.Port(viper.GetInt("port")),
				server.MaxUploadSize(viper.GetInt64("max_upload_size")),
				server.RateLimit(viper.GetInt("rate_limit")),
			}

			switch storageProvider := viper.GetString("storage"); storageProvider {
			case "local":
				if udir := viper.GetString("upload_directory"); udir == "" {
					log.Println("Upload directory must be set for local storage!")
					os.Exit(0)
				} else if storage, err := storage.NewLocalStorage(udir); err != nil {
					log.Println(err)
					os.Exit(1)
				} else {
					serverOptions = append(serverOptions, server.UseStorage(storage))
				}
			case "s3":
				bucket := viper.GetString("s3_bucket")
				region := viper.GetString("s3_region")
				accessKey := viper.GetString("aws_access_key")
				secretKey := viper.GetString("aws_secret_key")
				sessionToken := viper.GetString("aws_session_token")
				endpointUrl := viper.GetString("aws_endpoint_url")

				if storage, err := storage.NewS3Storage(accessKey, secretKey, sessionToken, endpointUrl, region, bucket); err != nil {
					log.Println(err)
					os.Exit(1)
				} else {
					serverOptions = append(serverOptions, server.UseStorage(storage))
				}
			case "storj":
				bucket := viper.GetString("storj_bucket")
				access := viper.GetString("storj_access")

				if storage, err := storage.NewStorjStorage(access, bucket); err != nil {
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
