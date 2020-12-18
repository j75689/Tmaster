package choice

import (
	"encoding/json"
	"fmt"

	"github.com/j75689/Tmaster/pkg/errors"
	"github.com/j75689/Tmaster/pkg/executor"
	"github.com/j75689/Tmaster/pkg/executor/choice/helper"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"github.com/j75689/Tmaster/pkg/message"
	"github.com/j75689/Tmaster/pkg/utils/parser"
	"github.com/rs/zerolog"
)

var _ executor.Executor = (*ChoiceExecutor)(nil)

func NewChoiceExecutor(logger zerolog.Logger) *ChoiceExecutor {
	return &ChoiceExecutor{
		logger: logger,
		helper: map[model.ChoiceHelper]helper.Helper{
			model.ChoiceHelperStringEquals:           helper.NewStringEq(),
			model.ChoiceHelperFloatEquals:            helper.NewFloatEq(),
			model.ChoiceHelperFloatGreaterThan:       helper.NewFloatGt(),
			model.ChoiceHelperFloatGreaterThanEquals: helper.NewFloatGte(),
			model.ChoiceHelperFloatLessThan:          helper.NewFloatLt(),
			model.ChoiceHelperFloatLessThanEquals:    helper.NewFloatLte(),
			model.ChoiceHelperIntEquals:              helper.NewIntEq(),
			model.ChoiceHelperIntGreaterThan:         helper.NewIntGt(),
			model.ChoiceHelperIntGreaterThanEquals:   helper.NewIntGte(),
			model.ChoiceHelperIntLessThan:            helper.NewIntLt(),
			model.ChoiceHelperIntLessThanEquals:      helper.NewIntLte(),
		},
	}
}

type ChoiceExecutor struct {
	logger zerolog.Logger
	helper map[model.ChoiceHelper]helper.Helper
}

func (executor *ChoiceExecutor) Execute(context message.Context, input interface{}, taskConfig *model.Task) (interface{}, *model.Task, errors.Error) {
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

	taskConfigStr, err := json.Marshal(taskConfig)
	if err != nil {
		return nil, taskConfig, errors.NewRuntimeError(fmt.Errorf("json marshal task config error [%v]", err))
	}
	taskConfigStr, err = parser.ReplaceVariables(taskConfigStr, inputParam)
	if err != nil {
		return nil, taskConfig, errors.NewRuntimeError(fmt.Errorf("replace variables of task config error [%v]", err))
	}
	taskConfigStr, err = parser.ReplaceSystemVariables(taskConfigStr, context)
	if err != nil {
		return nil, taskConfig, errors.NewRuntimeError(fmt.Errorf("replace system variables of task config error [%v]", err))
	}
	err = json.Unmarshal(taskConfigStr, taskConfig)
	if err != nil {
		return nil, taskConfig, errors.NewRuntimeError(fmt.Errorf("json unmarshal task config error [%v]", err))
	}
	executor.logger.Debug().Bytes("task_config", taskConfigStr).Msg("replaced task config")

	// default choice
	if taskConfig.Default != nil {
		taskConfig.Next = taskConfig.Default
	}

	// choice next
	for _, choice := range taskConfig.Choices {
		var (
			logic = false
			err   error
		)
		// normal
		if choice.Variable != nil && choice.Helper != nil {
			variable := *choice.Variable
			helperName := *choice.Helper
			if helperFunc, ok := executor.helper[helperName]; ok {
				logic, err = helperFunc(variable, choice.String, choice.Int, choice.Float)
				if err != nil {
					return nil, taskConfig, errors.NewRuntimeError(fmt.Errorf("choice helper error [%v]", err))
				}
			}
		}

		// not
		if choice.Not != nil {
			if helperFunc, ok := executor.helper[choice.Not.Helper]; ok {
				logic, err = helperFunc(choice.Not.Variable, choice.Not.String, choice.Not.Int, choice.Not.Float)
				if err != nil {
					return nil, taskConfig, errors.NewRuntimeError(fmt.Errorf("choice helper error [%v]", err))
				}
				logic = !logic
			}
		}

		// and
		for _, choiceHelper := range choice.And {
			if helperFunc, ok := executor.helper[choiceHelper.Helper]; ok {
				andlogic, err := helperFunc(choiceHelper.Variable, choiceHelper.String, choiceHelper.Int, choiceHelper.Float)
				if err != nil {
					return nil, taskConfig, errors.NewRuntimeError(fmt.Errorf("choice helper error [%v]", err))
				}
				logic = andlogic && logic
			}
		}

		// or
		for _, choiceHelper := range choice.Or {
			if helperFunc, ok := executor.helper[choiceHelper.Helper]; ok {
				andlogic, err := helperFunc(choiceHelper.Variable, choiceHelper.String, choiceHelper.Int, choiceHelper.Float)
				if err != nil {
					return nil, taskConfig, errors.NewRuntimeError(fmt.Errorf("choice helper error [%v]", err))
				}
				logic = andlogic || logic
			}
		}

		// match
		if logic {
			if choice.Next != nil {
				taskConfig.Next = choice.Next
			}

			return output, taskConfig, nil
		}
	}

	return output, taskConfig, nil
}
