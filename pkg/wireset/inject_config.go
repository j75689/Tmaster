package wireset

import (
	"github.com/google/wire"
	"github.com/j75689/Tmaster/pkg/config"
)

var ConfigSet = wire.NewSet(
	InitializeConfig,
)

func InitializeConfig(configPath string) (cfg config.Config, err error) {
	return config.NewConfig(configPath)
}
