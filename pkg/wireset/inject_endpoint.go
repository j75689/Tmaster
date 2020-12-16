package wireset

import (
	"github.com/google/wire"
	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/endpoint/grpc"
	"github.com/j75689/Tmaster/pkg/endpoint/http"
	"github.com/j75689/Tmaster/pkg/endpoint/nats"
	"github.com/j75689/Tmaster/pkg/endpoint/pubsub"
	redisstream "github.com/j75689/Tmaster/pkg/endpoint/redis_stream"
)

var EndpointSet = wire.NewSet(
	InitializeHttpHandler,
	InitializeGrpcHandler,
	InitializePubSubHandler,
	InitializeNatsHandler,
	InitializeRedisStreamHandler,
)

func InitializeHttpHandler(logger zerolog.Logger) *http.HttpHandler {
	return http.NewHttpHandler(logger)
}

func InitializeGrpcHandler(logger zerolog.Logger) *grpc.GrpcHandler {
	return grpc.NewGrpcHandler(logger)
}

func InitializePubSubHandler(logger zerolog.Logger) *pubsub.PubSubHandler {
	return pubsub.NewPubSubHandler(logger)
}

func InitializeNatsHandler(logger zerolog.Logger) *nats.NatsHandler {
	return nats.NewNatsHandler(logger)
}

func InitializeRedisStreamHandler(logger zerolog.Logger) *redisstream.RedisStreamHandler {
	return redisstream.NewRedisStreamHandler(logger)
}
