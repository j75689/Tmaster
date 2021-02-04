package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/j75689/Tmaster/pkg/utils/launcher"
	"github.com/j75689/Tmaster/service/dbhelper"
	"github.com/spf13/cobra"
)

var (
	dbHelperCmd = &cobra.Command{
		Use:           "dbhelper",
		Short:         "Start db helper",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(_ *cobra.Command, _ []string) {
			app, err := dbhelper.Initialize(cfgFile)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			launcher.Launch(app.Start, app.Shutdown, time.Duration(timeout)*time.Second)
		},
	}
)
