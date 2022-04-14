package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "development"

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show current version",
		Long:  "Show current program version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("fileigloo %s\n", Version)
		},
	}
)
