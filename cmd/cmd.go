package cmd

import (
	"fmt"
	"log"

	colors "github.com/logrusorgru/aurora"
	"github.com/urfave/cli/v2"

	"github.com/exler/fileigloo/server"
)

type Cmd struct {
	*cli.App
}

var Version = "0.1.0"

var globalFlags = []cli.Flag{}

var cliCommands = []*cli.Command{
	{
		Name:  "run",
		Usage: "Run the webapp",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "Server port number",
				Value:   "8000",
			},
			&cli.StringFlag{
				Name:    "upload-directory",
				Aliases: []string{"d"},
				Usage:   "Directory to upload files to",
				Value:   "uploads/",
			},
			&cli.Int64Flag{
				Name:    "max-upload-size",
				Aliases: []string{"s"},
				Usage:   "Max upload size",
				Value:   10000,
			},
			&cli.IntFlag{
				Name:    "rate-limit",
				Aliases: []string{"r"},
				Usage:   "Max allowed requests per second",
				Value:   5,
			},
			&cli.IntFlag{
				Name:  "purge-older",
				Usage: "How long before uploaded files are deleted (in hours)",
				Value: 24,
			},
			&cli.IntFlag{
				Name:  "purge-interval",
				Usage: "How often to check for expired files (in hours) - 0 disables purging",
				Value: 24,
			},
			&cli.StringFlag{
				Name:  "storage",
				Usage: "Choices: local",
				Value: "local",
			},
		},
		Action: func(c *cli.Context) error {
			var serverOptions = []server.OptionFn{}

			if c.String("port") != "" {
				serverOptions = append(serverOptions, server.Port(c.String("port")))
			}
			if c.Int64("max-upload-size") != 0 {
				serverOptions = append(serverOptions, server.MaxUploadSize(c.Int64("max-upload-size")))
			}
			if c.Int("rate-limit") != 0 {
				serverOptions = append(serverOptions, server.RateLimit(c.Int("rate-limit")))
			}
			if !(c.Int("purge-older") < 0) && !(c.Int("purge-interval") < 0) {
				serverOptions = append(serverOptions, server.Purge(c.Int("purge-older"), c.Int("purge-interval")))
			}
			switch storageProvider := c.String("storage-provider"); storageProvider {
			case "local":
				if udir := c.String("upload-directory"); udir == "" {
					panic("Upload directory must be set for local storage!")
				} else if storage, err := server.NewLocalStorage(udir); err != nil {
					return err
				} else {
					serverOptions = append(serverOptions, server.UseStorage(storage))
				}
			}

			srv := server.New(serverOptions...)
			err := srv.Run()

			if err != nil {
				return err
			}

			return nil
		},
	},
	{
		Name:    "version",
		Usage:   "Show current version",
		Aliases: []string{"v"},
		Action: func(c *cli.Context) error {
			fmt.Printf("fileigloo %s", Version)
			return nil
		},
	},
}

func New() *Cmd {
	log.SetPrefix(colors.Blue("[fileigloo] ").String())

	app := &cli.App{
		Name:     fmt.Sprintf("fileigloo %s", Version),
		Usage:    "Small and simple temporary file sharing & pastebin",
		Flags:    globalFlags,
		Commands: cliCommands,
	}

	return &Cmd{
		App: app,
	}
}
