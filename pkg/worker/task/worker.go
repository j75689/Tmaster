package task

import (
	"context"
	"fmt"
	"time"

	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/errors"
	"github.com/j75689/Tmaster/pkg/executor"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"github.com/j75689/Tmaster/pkg/lock"
	"github.com/j75689/Tmaster/pkg/message"
	"github.com/j75689/Tmaster/pkg/opentracer"
	"github.com/rs/zerolog"
)

const (
	_WorkerName = "Task"
)

func NewWorker(
	config config.Config,
	taskExecutor,
	parallelExecutor,
	waitExecutor,
	passExecutor,
	choiceExecutor executor.Executor,
	lock lock.Locker,
	logger zerolog.Logger,
	lockTimeout time.Duration,
	tracer *opentracer.ServiceTracer,
) *TaskWorker {
	return &TaskWorker{
		config: config,
		executor: map[model.TaskType]executor.Executor{
			model.TaskTypeTask:     taskExecutor,
			model.TaskTypeParallel: parallelExecutor,
			model.TaskTypePass:     passExecutor,
			model.TaskTypeWait:     waitExecutor,
			model.TaskTypeChoice:   choiceExecutor,
		},
		lock:        lock,
		logger:      logger,
		lockTimeout: lockTimeout,
		tracer:      tracer,
	}
}

type TaskWorker struct {
	config      config.Config
	executor    map[model.TaskType]executor.Executor
	lock        lock.Locker
	logger      zerolog.Logger
	lockTimeout time.Duration
	tracer      *opentracer.ServiceTracer
}

func (worker *TaskWorker) Process(taskInput *message.TaskInput) (*message.TaskOutput, error) {
	locked, _ := worker.lock.LockWithAutoDelay(taskInput.Context.Context, taskInput.TaskXID)
	if !locked {
		return nil, fmt.Errorf("duplicate lock for %s", taskInput.TaskXID)
	}
	defer worker.lock.UnLock(taskInput.Context.Context, taskInput.TaskXID)

	if worker.config.OpenTracing.Enable {
		var (
			traceRecord *opentracer.TraceRecord
			terr        error
			taskName    = taskInput.Task.Name
			tags        = map[string]interface{}{"job_id": taskInput.Context.Job.JobID}
		)

		if taskInput.Context.TraceCarrier != nil {
			traceRecord, terr = worker.tracer.TraceServer(
				taskName,
				taskInput.Context.TraceCarrier,
				tags,
			)
		} else {
			traceRecord, terr = worker.tracer.TraceClient(
				taskName,
				tags,
			)
		}
		if terr != nil {
			worker.logger.Err(terr).Msg("create trace error")
		}
		defer traceRecord.Finish()
		taskInput.Context.TraceCarrier = traceRecord.Carrier()
	}

	taskOutput := &message.TaskOutput{
		Context: message.Context{
			TraceCarrier: taskInput.Context.TraceCarrier,
			Execution: message.Execution{
				ID:                taskInput.Context.Execution.ID,
				StartTime:         time.Now(),
				Cause:             taskInput.Context.Execution.Cause,
				MaxConsistentNums: taskInput.Context.Execution.MaxConsistentNums,
				ConsistentNums:    taskInput.Context.Execution.ConsistentNums,
				Timeout:           taskInput.Context.Execution.Timeout,
				MaxTaskExecution:  taskInput.Context.Execution.MaxTaskExecution,
				TaskExecution:     taskInput.Context.Execution.TaskExecution,
			},
			State: message.State{
				EnteredTime: taskInput.Context.State.EnteredTime,
				Name:        taskInput.Context.State.Name,
				RetryCount:  taskInput.Context.State.RetryCount,
				Status:      model.StatusSuccess,
			},
			Job:   taskInput.Context.Job,
			Tasks: taskInput.Context.Tasks,
		},
		From:       taskInput.From,
		TaskXID:    taskInput.TaskXID,
		Task:       taskInput.Task,
		InputValue: taskInput.InputValue,
	}
	executor := worker.executor[taskInput.Task.Type]
	if executor == nil {
		err := errors.NewRuntimeError(fmt.Errorf("[%s] task type is not existed", taskInput.Task.Type))
		code := err.ErrCode()
		message := err.Error()
		taskOutput.ErrorCode = &code
		taskOutput.Error = &message
		taskOutput.Context.Execution.EndTime = time.Now()
		taskOutput.Context.State.Status = model.StatusFailed
		return taskOutput, nil
	}

	taskInput.Context.Context = context.Background()
	if taskInput.Task.Timeout != nil {
		timeout := time.Duration(*taskInput.Task.Timeout) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		taskInput.Context.Context = ctx
	}

	outputValue, outputTask, err := executor.Execute(taskInput.Context, taskInput.InputValue, &taskInput.Task)
	if err != nil {
		code := err.ErrCode()
		message := err.Error()
		taskOutput.ErrorCode = &code
		taskOutput.Error = &message
		taskOutput.Context.State.Status = model.StatusFailed
	}
	if outputTask != nil {
		taskOutput.Task = *outputTask
	}

	taskOutput.OutputValue = outputValue
	taskOutput.Context.Execution.EndTime = time.Now()
	return taskOutput, nil
}
