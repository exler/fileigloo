package cmd

import (
	"log"

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
				server.Port(viper.GetInt("Port")),
				server.MaxUploadSize(viper.GetInt64("max_upload_size")),
				server.RateLimit(viper.GetInt("rate_limit")),
			}

			if viper.GetBool("https") {
				if domains := viper.GetStringSlice("domain"); len(domains) == 0 {
					log.Fatalln("Please pass domains for the SSL certificate as arguments (usage: --domain='example.com')")
				} else {
					serverOptions = append(serverOptions, server.HTTPS(domains))
				}
			}

			switch storageProvider := viper.GetString("storage"); storageProvider {
			case "local":
				if udir := viper.GetString("upload_directory"); udir == "" {
					log.Fatalln("Upload directory must be set for local storage!")
				} else if storage, err := storage.NewLocalStorage(udir, viper.GetInt("purge_interval"), viper.GetInt("purge_older")); err != nil {
					log.Fatalln(err)
				} else {
					serverOptions = append(serverOptions, server.UseStorage(storage))
				}
			case "s3":
				bucket := viper.GetString("s3_bucket")
				region := viper.GetString("s3_region")
				accessKey := viper.GetString("aws_access_key")
				secretKey := viper.GetString("aws_secret_key")
				sessionToken := viper.GetString("aws_session_token")

				if storage, err := storage.NewS3Storage(accessKey, secretKey, sessionToken, region, bucket); err != nil {
					log.Fatalln(err)
				} else {
					serverOptions = append(serverOptions, server.UseStorage(storage))
				}
			default:
				log.Fatalln("Incorrect or no storage type chosen!")
			}

			srv := server.New(serverOptions...)
			srv.Run()
		},
	}
)

func init() {
	serverCmd.Flags().Int("port", 8000, "Port to run the server on")
	serverCmd.Flags().Bool("https", false, "Enable HTTPS server (requires --domain to be set)")
	serverCmd.Flags().StringSlice("domain", []string{}, "Domains for Let's Encrypt certificate")
	viper.BindPFlag("port", serverCmd.Flags().Lookup("port"))
	viper.BindPFlag("https", serverCmd.Flags().Lookup("https"))
	viper.BindPFlag("domain", serverCmd.Flags().Lookup("domain"))
}
