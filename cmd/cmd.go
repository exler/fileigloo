package cmd

import (
	"log"

	. "github.com/logrusorgru/aurora"
	"github.com/urfave/cli/v2"
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
			log.Printf("Running webapp on port %s", port)
			return nil
		},
	},
}

func New() *Cmd {
	log.SetPrefix(Blue("[fileigloo] ").String())

	app := &cli.App{
		Name:     "fileigloo",
		Usage:    "Exchange files",
		Flags:    globalFlags,
		Commands: cliCommands,
	}

	return &Cmd{
		App: app,
	}
}
