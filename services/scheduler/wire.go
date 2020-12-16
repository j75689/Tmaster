//+build wireinject

//The build tag makes sure the stub is not built in the final build.

package scheduler

import (
	"github.com/google/wire"
	"github.com/j75689/Tmaster/pkg/wireset"
)

func Initialize(configPath string) (Application, error) {
	wire.Build(
		newApplication,
		wireset.WorkerSet,
		wireset.TracerSet,
		wireset.LockerSet,
		wireset.DatabaseSet,
		wireset.MQSet,
		wireset.LoggerSet,
		wireset.ConfigSet,
	)
	return Application{}, nil
}
