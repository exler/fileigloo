package cmd

import (
	"os"

	"github.com/urfave/cli/v2"
)

var Cmd = &cli.App{
	Name:     "fileigloo",
	Usage:    "Small and simple online file sharing & pastebin",
	Commands: []*cli.Command{versionCmd, serverCmd},
}

func Run() error {
	return Cmd.Run(os.Args)
}
