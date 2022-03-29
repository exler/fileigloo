package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "0.3.1"

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show current version",
		Long:  "Show current program version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("fileigloo %s\n", Version)
		},
	}
)
