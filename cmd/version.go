package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version    string
	commitID   string
	commitDate string
)

var (
	versionCmd = &cobra.Command{
		Use:          "version",
		Short:        "Print version information",
		SilenceUsage: true,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("Version: %s\n", version)
			fmt.Printf("Commit ID: %s\n", commitID)
			fmt.Printf("Commit Date: %s\n", commitDate)
		},
	}
)
