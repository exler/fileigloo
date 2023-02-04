package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var (
	Version = "development"

	versionCmd = &cli.Command{
		Name:  "version",
		Usage: "Show current version",
		Action: func(cCtx *cli.Context) error {
			fmt.Printf("fileigloo %s\n", Version)
			return nil
		},
	}
)
