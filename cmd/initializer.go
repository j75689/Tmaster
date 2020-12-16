package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/j75689/Tmaster/pkg/utils/launcher"
	"github.com/j75689/Tmaster/services/initializer"
)

var (
	initialzierCmd = &cobra.Command{
		Use:           "initializer",
		Short:         "Start job initializer",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(_ *cobra.Command, _ []string) {
			app, err := initializer.Initialize(cfgFile)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			launcher.Launch(app.Start, app.Shutdown, time.Duration(timeout)*time.Second)
		},
	}
)
