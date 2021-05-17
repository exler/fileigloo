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

var port string

var globalFlags = []cli.Flag{}

var cliCommands = []*cli.Command{
	{
		Name:  "run",
		Usage: "Run the webapp",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "port",
				Aliases:     []string{"p"},
				Usage:       "Port number of the webapp",
				Value:       "8000",
				Destination: &port,
			},
		},
		Action: func(c *cli.Context) error {
			srv := server.New()
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
