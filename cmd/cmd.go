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

var Version = "0.0.1"

var globalFlags = []cli.Flag{}

var cliCommands = []*cli.Command{
	{
		Name:  "run",
		Usage: "Run the webapp",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "Port number of the webapp",
				Value:   "8000",
			},
			&cli.StringFlag{
				Name:    "upload-directory",
				Aliases: []string{"d"},
				Usage:   "Directory to upload to",
				Value:   "uploads/",
			},
			&cli.Int64Flag{
				Name:    "max-upload-size",
				Aliases: []string{"s"},
				Usage:   "Max upload size",
				Value:   10000,
			},
		},
		Action: func(c *cli.Context) error {
			var serverOptions = []server.OptionFn{}

			if c.String("port") != "" {
				serverOptions = append(serverOptions, server.Port(c.String("port")))
			}
			if c.String("upload-directory") != "" {
				serverOptions = append(serverOptions, server.UploadDirectory(c.String("upload-directory")))
			}
			if c.Int64("max-upload-size") != 0 {
				serverOptions = append(serverOptions, server.MaxUploadSize(c.Int64("max-upload-size")))
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
		Usage:    "Exchange files",
		Flags:    globalFlags,
		Commands: cliCommands,
	}

	return &Cmd{
		App: app,
	}
}
