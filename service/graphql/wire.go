//+build wireinject

//The build tag makes sure the stub is not built in the final build.

package graphql

import (
	"github.com/google/wire"
	"github.com/j75689/Tmaster/pkg/wireset"
)

func Initialize(configPath string) (Application, error) {
	wire.Build(
		newApplication,
		wireset.HttpSet,
		wireset.TracerSet,
		wireset.GraphqlSchemaSet,
		wireset.DatabaseSet,
		wireset.MQSet,
		wireset.LoggerSet,
		wireset.ConfigSet,
	)
	return Application{}, nil
}
