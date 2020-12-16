package wireset

import (
	"github.com/google/wire"
	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/logger"
)

var LoggerSet = wire.NewSet(
	InitializeLogger,
)

func InitializeLogger(config config.Config) (zerolog.Logger, error) {
	return logger.NewLogger(config.LogLevel, config.LogFormat)
}
