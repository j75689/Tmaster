package http

import (
	"encoding/json"
	"fmt"

	"github.com/j75689/Tmaster/pkg/endpoint"
	"github.com/j75689/Tmaster/pkg/endpoint/grpc"
	"github.com/j75689/Tmaster/pkg/endpoint/http"
	"github.com/j75689/Tmaster/pkg/endpoint/nats"
	"github.com/j75689/Tmaster/pkg/endpoint/pubsub"
	redisstream "github.com/j75689/Tmaster/pkg/endpoint/redis_stream"
	"github.com/j75689/Tmaster/pkg/errors"
	"github.com/j75689/Tmaster/pkg/executor"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"github.com/j75689/Tmaster/pkg/message"
	"github.com/j75689/Tmaster/pkg/utils/parser"
	"github.com/rs/zerolog"
)

var _ executor.Executor = (*TaskExecutor)(nil)

func NewTaskExecutor(
	httpHandler *http.HttpHandler,
	grpcHandler *grpc.GrpcHandler,
	pubsubHandler *pubsub.PubSubHandler,
	natsHandler *nats.NatsHandler,
	redisstream *redisstream.RedisStreamHandler,
	logger zerolog.Logger,
) *TaskExecutor {
	endpointHandlers := map[model.Protocol]endpoint.Handler{
		model.ProtocolHTTP:        httpHandler,
		model.ProtocolGrpc:        grpcHandler,
		model.ProtocolPubsub:      pubsubHandler,
		model.ProtocolNats:        natsHandler,
		model.ProtocolRedisStream: redisstream,
	}
	return &TaskExecutor{
		logger:           logger,
		endpointHandlers: endpointHandlers,
	}
}

type TaskExecutor struct {
	logger           zerolog.Logger
	endpointHandlers map[model.Protocol]endpoint.Handler
}

type TaskResult struct {
	value interface{}
	err   errors.Error
}

func (executor *TaskExecutor) Execute(context message.Context, input interface{}, taskConfig *model.Task) (interface{}, *model.Task, errors.Error) {
	resultchan := make(chan *TaskResult)
	go func() {
		result := &TaskResult{}
		defer close(resultchan)
		// get input
		var err error
		inputParam := input
		if taskConfig.InputPath != nil {
			inputPath := *taskConfig.InputPath
			inputParam, err = parser.GetJSONValue(inputPath, input)
			if err != nil {
				result.err = errors.NewRuntimeError(fmt.Errorf("get input value error [%v]", err))
				return
			}
		}
		outputParam := inputParam

		// always returns output value
		defer func() {
			// get output
			output := outputParam
			if taskConfig.OutputPath != nil {
				outputPath := *taskConfig.OutputPath
				output, err = parser.GetJSONValue(outputPath, outputParam)
				if err != nil {
					result.err = errors.NewRuntimeError(fmt.Errorf("get output value error [%v]", err))
				}
			}
			result.value = output
			resultchan <- result
		}()

		// execute
		handler := executor.endpointHandlers[taskConfig.Endpoint.Protocol]
		if handler == nil {
			result.err = errors.NewRuntimeError(fmt.Errorf("endpoint handler [%s] not found", taskConfig.Endpoint.Protocol))
			return
		}
		endpointConfig := taskConfig.Endpoint
		endpointConfigStr, err := json.Marshal(endpointConfig)
		if err != nil {
			result.err = errors.NewRuntimeError(fmt.Errorf("json marshal endpoint config error [%v]", err))
			return
		}
		endpointConfigStr, err = parser.ReplaceVariables(endpointConfigStr, inputParam)
		if err != nil {
			result.err = errors.NewRuntimeError(fmt.Errorf("replace variables of endpoint config error [%v]", err))
			return
		}
		endpointConfigStr, err = parser.ReplaceSystemVariables(endpointConfigStr, context)
		if err != nil {
			result.err = errors.NewRuntimeError(fmt.Errorf("replace system variables of endpoint config error [%v]", err))
			return
		}
		err = json.Unmarshal(endpointConfigStr, endpointConfig)
		if err != nil {
			result.err = errors.NewRuntimeError(fmt.Errorf("json unmarshal endpoint config error [%v]", err))
			return
		}
		headerResult, handlerResult, err := handler.Do(context.Context, endpointConfig)
		if err != nil {
			result.err = errors.NewTaskFailedError(err)

			// set error message
			if taskConfig.ErrorPath != nil {
				errorPath := *taskConfig.ErrorPath
				outputParam, err = parser.SetJSONValue(errorPath, err.Error(), inputParam)
				if err != nil {
					result.err = errors.NewRuntimeError(fmt.Errorf("set error value error [%v]", err))
					return
				}
			}
		}

		// set result
		if taskConfig.ResultPath != nil {
			resultPath := *taskConfig.ResultPath
			outputParam, err = parser.SetJSONValue(resultPath, handlerResult, inputParam)
			if err != nil {
				result.err = errors.NewRuntimeError(fmt.Errorf("set result value error [%v]", err))
				return
			}
		}

		// set header result
		if taskConfig.HeaderPath != nil {
			headerPath := *taskConfig.HeaderPath
			outputParam, err = parser.SetJSONValue(headerPath, headerResult, inputParam)
			if err != nil {
				result.err = errors.NewRuntimeError(fmt.Errorf("set header result value error [%v]", err))
				return
			}
		}
	}()

	select {
	case <-context.Done():
		return nil, taskConfig, errors.NewTimeoutError(context.Err())
	case result := <-resultchan:
		return result.value, taskConfig, result.err
	}
}
