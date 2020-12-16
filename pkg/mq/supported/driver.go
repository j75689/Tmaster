package supported

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/j75689/Tmaster/pkg/mq/nats"
	"github.com/j75689/Tmaster/pkg/mq/pubsub"
	redisstream "github.com/j75689/Tmaster/pkg/mq/redis_stream"
)

type MQDriver string

func (driver MQDriver) String() string {
	return string(driver)
}

const (
	GooglePubSub MQDriver = "google_pub_sub"
	Nats         MQDriver = "nats"
	RedisStream  MQDriver = "redis_stream"
)

var supported = map[MQDriver]func(
	config config.MQConfig,
	logger zerolog.Logger,
) (mq.MQ, error){
	GooglePubSub: pubsub.NewGooglePubSub,
	Nats:         nats.NewNats,
	RedisStream:  redisstream.NewRedisStream,
}

func NewMQDriver(config config.MQConfig, logger zerolog.Logger) (mq.MQ, error) {
	if f, ok := supported[MQDriver(config.Driver)]; ok {
		return f(config, logger)
	}
	return nil, fmt.Errorf("unsupported mq driver [%s]", config.Driver)
}
