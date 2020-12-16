package pass

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/errors"
	"github.com/j75689/Tmaster/pkg/executor"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"github.com/j75689/Tmaster/pkg/message"
	"github.com/j75689/Tmaster/pkg/utils/parser"
)

var _ executor.Executor = (*PassExecutor)(nil)

func NewPassExecutorr(logger zerolog.Logger) *PassExecutor {
	return &PassExecutor{
		logger: logger,
	}
}

type PassExecutor struct {
	logger zerolog.Logger
}

func (executor *PassExecutor) Execute(context message.Context, input interface{}, taskConfig *model.Task) (interface{}, *model.Task, errors.Error) {
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
