package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/j75689/Tmaster/pkg/utils/launcher"
	"github.com/j75689/Tmaster/service/initmq"
	"github.com/spf13/cobra"
)

var (
	initMQCmd = &cobra.Command{
		Use:           "initmq",
		Short:         "Initialize MQ",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	initGraphqlMQ = &cobra.Command{
		Use:           "graphql",
		Short:         "Start Initialize MQ Of Graphql",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(_ *cobra.Command, _ []string) {
			app, err := initmq.Initialize(cfgFile)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			launcher.Launch(app.InitGraphqlMQ, nil, time.Duration(timeout)*time.Second)
		},
	}

	initInitializerMQ = &cobra.Command{
		Use:           "initializer",
		Short:         "Start Initialize MQ Of Initializer",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(_ *cobra.Command, _ []string) {
			app, err := initmq.Initialize(cfgFile)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			launcher.Launch(app.InitInitializerMQ, nil, time.Duration(timeout)*time.Second)
		},
	}

	initSchedulerMQ = &cobra.Command{
		Use:           "scheduler",
		Short:         "Start Initialize MQ Of Scheduler",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(_ *cobra.Command, _ []string) {
			app, err := initmq.Initialize(cfgFile)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			launcher.Launch(app.InitSchedulerMQ, nil, time.Duration(timeout)*time.Second)
		},
	}

	initWorkerMQ = &cobra.Command{
		Use:           "worker",
		Short:         "Start Initialize MQ Of Worker",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(_ *cobra.Command, _ []string) {
			app, err := initmq.Initialize(cfgFile)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			launcher.Launch(app.InitWorkerMQ, nil, time.Duration(timeout)*time.Second)
		},
	}

	initDBHelperMQ = &cobra.Command{
		Use:           "dbhelper",
		Short:         "Start Initialize MQ Of DBHelper",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(_ *cobra.Command, _ []string) {
			app, err := initmq.Initialize(cfgFile)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			launcher.Launch(app.InitDBHelperMQ, nil, time.Duration(timeout)*time.Second)
		},
	}
)

func init() {
	initMQCmd.AddCommand(initGraphqlMQ, initInitializerMQ, initSchedulerMQ, initWorkerMQ, initDBHelperMQ)
}
