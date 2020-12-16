package wireset

import (
	"github.com/google/wire"
	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/endpoint/grpc"
	"github.com/j75689/Tmaster/pkg/endpoint/http"
	"github.com/j75689/Tmaster/pkg/endpoint/nats"
	"github.com/j75689/Tmaster/pkg/endpoint/pubsub"
	redisstream "github.com/j75689/Tmaster/pkg/endpoint/redis_stream"

	choiceExecutor "github.com/j75689/Tmaster/pkg/executor/choice"
	parallelExecutor "github.com/j75689/Tmaster/pkg/executor/parallel"
	passExecutor "github.com/j75689/Tmaster/pkg/executor/pass"
	taskExecutor "github.com/j75689/Tmaster/pkg/executor/task"
	waitExecutor "github.com/j75689/Tmaster/pkg/executor/wait"
	"github.com/j75689/Tmaster/pkg/mq"
)

var ExecutorSet = wire.NewSet(
	InitializeParallelExecutor,
	InitializePassExecutor,
	InitializeTaskExecutor,
	InitializeWaitExecutor,
	InitializeChoiceExecutor,
)

func InitializeTaskExecutor(
	http *http.HttpHandler,
	grpc *grpc.GrpcHandler,
	pubsub *pubsub.PubSubHandler,
	nats *nats.NatsHandler,
	redisstream *redisstream.RedisStreamHandler,
	logger zerolog.Logger,
) *taskExecutor.TaskExecutor {
	return taskExecutor.NewTaskExecutor(http, grpc, pubsub, nats, redisstream, logger)
}

func InitializeParallelExecutor(
	config config.Config,
	mq mq.MQ,
	logger zerolog.Logger,
) *parallelExecutor.ParallelExecutor {
	return parallelExecutor.NewParallelExecutor(config.TaskWorker, config.MQ, mq, logger)
}

func InitializeWaitExecutor(
	logger zerolog.Logger,
) *waitExecutor.WaitExecutor {
	return waitExecutor.NewWaitExecutor(logger)
}

func InitializePassExecutor(
	logger zerolog.Logger,
) *passExecutor.PassExecutor {
	return passExecutor.NewPassExecutorr(logger)
}

func InitializeChoiceExecutor(
	logger zerolog.Logger,
) *choiceExecutor.ChoiceExecutor {
	return choiceExecutor.NewChoiceExecutor(logger)
}
