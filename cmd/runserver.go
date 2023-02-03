package cmd

import (
	"log"
	"os"

	"github.com/exler/fileigloo/logger"
	"github.com/exler/fileigloo/server"
	"github.com/exler/fileigloo/storage"
	"github.com/urfave/cli/v2"
)

var serverCmd = &cli.Command{
	Name:  "runserver",
	Usage: "Run web server",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Usage:   "Port to listen on",
			Value:   8000,
		},
		&cli.Int64Flag{
			Name:    "max-upload-size",
			Value:   0,
			EnvVars: []string{"MAX_UPLOAD_SIZE"},
		},
		&cli.IntFlag{
			Name:    "rate-limit",
			Value:   2,
			EnvVars: []string{"RATE_LIMIT"},
		},
		&cli.StringFlag{
			Name:    "storage",
			Value:   "local",
			EnvVars: []string{"STORAGE"},
		},
		&cli.StringFlag{
			Name:    "upload-directory",
			Value:   "uploads/",
			EnvVars: []string{"UPLOAD_DIRECTORY"},
		},
		&cli.StringFlag{
			Name:    "s3-bucket",
			EnvVars: []string{"S3_BUCKET"},
		},
		&cli.StringFlag{
			Name:    "s3-region",
			EnvVars: []string{"S3_REGION"},
		},
		&cli.StringFlag{
			Name:    "aws-access-key",
			EnvVars: []string{"AWS_ACCESS_KEY"},
		},
		&cli.StringFlag{
			Name:    "aws-secret-key",
			EnvVars: []string{"AWS_SECRET_KEY"},
		},
		&cli.StringFlag{
			Name:    "aws-session-token",
			EnvVars: []string{"AWS_SESSION_TOKEN"},
		},
		&cli.StringFlag{
			Name:    "aws-endpoint-url",
			EnvVars: []string{"AWS_ENDPOINT_URL"},
		},
		&cli.StringFlag{
			Name:    "sentry-dsn",
			EnvVars: []string{"SENTRY_DSN"},
		},
	},
	Action: func(cCtx *cli.Context) error {
		serverOptions := []server.OptionFn{
			server.Port(cCtx.Int("port")),
			server.MaxUploadSize(cCtx.Int64("max-upload-size")),
			server.RateLimit(cCtx.Int("rate-limit")),
			server.UseLogger(logger.NewLogger(cCtx.String("sentry-dsn"))),
		}

		switch storageProvider := cCtx.String("storage"); storageProvider {
		case "local":
			if udir := cCtx.String("upload-directory"); udir == "" {
				log.Println("Upload directory must be set for local storage!")
				os.Exit(0)
			} else if storage, err := storage.NewLocalStorage(udir); err != nil {
				log.Fatalln(err)
			} else {
				serverOptions = append(serverOptions, server.UseStorage(storage))
			}
		case "s3":
			bucket := cCtx.String("s3-bucket")
			region := cCtx.String("s3-region")
			accessKey := cCtx.String("aws-access-key")
			secretKey := cCtx.String("aws-secret-key")
			sessionToken := cCtx.String("aws-session-token")
			endpointUrl := cCtx.String("aws-endpoint-url")

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

		return nil
	},
}
