package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/j75689/Tmaster/pkg/utils/launcher"
	"github.com/j75689/Tmaster/service/migrate"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:           "migrate",
	Short:         "Start database migration",
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(_ *cobra.Command, _ []string) {
		app, err := migrate.Initialize(cfgFile)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		launcher.Launch(app.Start, nil, time.Duration(timeout)*time.Second)
	},
}
