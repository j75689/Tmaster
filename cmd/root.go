package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Root command
var (
	timeout uint
	cfgFile string
	rootCmd = &cobra.Command{
		SilenceUsage: true,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd, graphqlCmd, workerCmd, schedulerCmd, initialzierCmd, dbHelperCmd, migrateCmd, initMQCmd)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config/default.config.yaml", "config file")
	rootCmd.PersistentFlags().UintVar(&timeout, "timeout", 300, "graceful shutdown timeout (second)")
}
