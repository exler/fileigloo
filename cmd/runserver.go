package cmd

import (
	"log"

	"github.com/exler/fileigloo/logger"
	"github.com/exler/fileigloo/server"
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
			Value:   20,
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
		&cli.StringFlag{
			Name:    "sentry-environment",
			EnvVars: []string{"SENTRY_ENVIRONMENT"},
		},
	},
	Action: func(cCtx *cli.Context) error {
		serverOptions := []server.OptionFn{
			server.Port(cCtx.Int("port")),
			server.MaxUploadSize(cCtx.Int64("max-upload-size")),
			server.MaxRequests(cCtx.Int("rate-limit")),
			server.UseLogger(logger.NewLogger(cCtx.String("sentry-dsn"), cCtx.String("sentry-environment"))),
		}

		storage, err := GetStorage(cCtx)
		if err != nil {
			log.Fatalln(err)
		}
		serverOptions = append(serverOptions, server.UseStorage(storage))

		srv := server.New(serverOptions...)
		srv.Run()

		return nil
	},
}
