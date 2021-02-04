//+build wireinject

//The build tag makes sure the stub is not built in the final build.

package migrate

import (
	"github.com/google/wire"
	"github.com/j75689/Tmaster/pkg/wireset"
)

func Initialize(configPath string) (Application, error) {
	wire.Build(
		newApplication,
		wireset.DatabaseSet,
		wireset.LoggerSet,
		wireset.ConfigSet,
	)
	return Application{}, nil
}
