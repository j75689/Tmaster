package wireset

import (
	"github.com/google/wire"
	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/j75689/Tmaster/pkg/mq/supported"
)

var MQSet = wire.NewSet(
	InitializeMQ,
)

func InitializeMQ(config config.Config, logger zerolog.Logger) (mq.MQ, error) {
	return supported.NewMQDriver(config.MQ, logger)
}
