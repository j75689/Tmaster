package parallel

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/errors"
	"github.com/j75689/Tmaster/pkg/executor"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"github.com/j75689/Tmaster/pkg/message"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

var _ executor.Executor = (*ParallelExecutor)(nil)

func NewParallelExecutor(config config.TaskWorkerConfig, mqConfig config.MQConfig, mq mq.MQ, logger zerolog.Logger) *ParallelExecutor {
	return &ParallelExecutor{
		config:   config,
		mqConfig: mqConfig,
		mq:       mq,
		logger:   logger,
	}
}

type ParallelExecutor struct {
	config   config.TaskWorkerConfig
	mqConfig config.MQConfig
	mq       mq.MQ
	logger   zerolog.Logger
}

func (executor *ParallelExecutor) Execute(context message.Context, input interface{}, taskConfig *model.Task) (interface{}, *model.Task, errors.Error) {
	for _, job := range taskConfig.Branches {
		id := xid.New()

		data, err := json.Marshal(&message.InitJob{
			JobID:       id.String(),
			ParentJobID: &context.Job.JobID,
			ParentID:    &context.Job.ID,
			Job:         *job,
		})
		if err != nil {
			return nil, taskConfig, errors.NewRuntimeError(fmt.Errorf("json marshal initjob error [%v]", err))
		}
		if err = executor.mq.Publish(
			rand.Int63n(int64(executor.mqConfig.Distribution)),
			executor.config.InitJob.ProjectID,
			executor.config.InitJob.Topic,
			data,
		); err != nil {
			return nil, taskConfig, errors.NewTaskFailedError(fmt.Errorf("publish init sub job [%s] error [%v]", id.String(), err))
		}
	}

	return model.StatusSuccess, taskConfig, nil
}
