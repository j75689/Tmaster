package wait

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/errors"
	"github.com/j75689/Tmaster/pkg/executor"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"github.com/j75689/Tmaster/pkg/message"
	"github.com/j75689/Tmaster/pkg/utils/parser"
)

var _ executor.Executor = (*WaitExecutor)(nil)

func NewWaitExecutor(logger zerolog.Logger) *WaitExecutor {
	return &WaitExecutor{
		logger: logger,
	}
}

type WaitExecutor struct {
	logger zerolog.Logger
}

func (executor *WaitExecutor) Execute(context message.Context, input interface{}, taskConfig *model.Task) (interface{}, *model.Task, errors.Error) {
	// get input
	var err error
	inputParam := input
	if taskConfig.InputPath != nil {
		inputPath := *taskConfig.InputPath
		inputParam, err = parser.GetJSONValue(inputPath, input)
		if err != nil {
			return nil, taskConfig, errors.NewRuntimeError(fmt.Errorf("get input value error [%v]", err))
		}
	}

	if taskConfig.Seconds != nil {
		time.Sleep(time.Duration(*taskConfig.Seconds) * time.Second)
	}

	if taskConfig.Until != nil {
		t := *taskConfig.Until
		time.Sleep(time.Duration(t.UnixNano() - time.Now().UnixNano()))
	}

	// get output
	outputParam := inputParam
	output := outputParam
	if taskConfig.OutputPath != nil {
		outputPath := *taskConfig.OutputPath
		output, err = parser.GetJSONValue(outputPath, outputParam)
		if err != nil {
			return nil, taskConfig, errors.NewRuntimeError(fmt.Errorf("get output value error [%v]", err))
		}
	}

	return output, taskConfig, nil
}
