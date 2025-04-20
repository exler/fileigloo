package cmd

import (
	"log"

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
			Value:   100,
			EnvVars: []string{"RATE_LIMIT"},
		},
		&cli.StringFlag{
			Name:    "storage",
			Value:   "local",
			EnvVars: []string{"STORAGE"},
		},
		&cli.StringFlag{
			Name:    "site-password",
			Usage:   "Password to protect the site with",
			EnvVars: []string{"SITE_PASSWORD"},
		},
		&cli.StringFlag{
			Name:    "upload-directory",
			Value:   "uploads/",
			EnvVars: []string{"UPLOAD_DIRECTORY"},
		},
		&cli.StringFlag{
			Name:    "aws-s3-bucket",
			EnvVars: []string{"AWS_S3_BUCKET"},
		},
		&cli.StringFlag{
			Name:    "aws-s3-region",
			EnvVars: []string{"AWS_S3_REGION"},
		},
		&cli.StringFlag{
			Name:    "aws-s3-access-key",
			EnvVars: []string{"AWS_S3_ACCESS_KEY"},
		},
		&cli.StringFlag{
			Name:    "aws-s3-secret-key",
			EnvVars: []string{"AWS_S3_SECRET_KEY"},
		},
		&cli.StringFlag{
			Name:    "aws-s3-session-token",
			EnvVars: []string{"AWS_S3_SESSION_TOKEN"},
		},
		&cli.StringFlag{
			Name:    "aws-s3-endpoint-url",
			EnvVars: []string{"AWS_S3_ENDPOINT_URL"},
		},
		&cli.StringFlag{
			Name:    "sentry-dsn",
			EnvVars: []string{"SENTRY_DSN"},
		},
		&cli.StringFlag{
			Name:    "sentry-environment",
			Value:   "undefined",
			EnvVars: []string{"SENTRY_ENVIRONMENT"},
		},
		&cli.Float64Flag{
			Name:    "sentry-traces-sample-rate",
			Value:   0,
			EnvVars: []string{"SENTRY_TRACES_SAMPLE_RATE"},
		},
	},
	Action: func(cCtx *cli.Context) error {
		serverOptions := []server.OptionFn{
			server.Port(cCtx.Int("port")),
			server.MaxUploadSize(cCtx.Int64("max-upload-size")),
			server.MaxRequests(cCtx.Int("rate-limit")),
			server.Sentry(cCtx.String("sentry-dsn"), cCtx.String("sentry-environment"), cCtx.Float64("sentry-traces-sample-rate")),
			server.SitePassword(cCtx.String("site-password")),
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
