package scheduler

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/j75689/Tmaster/pkg/config"
	dbmodel "github.com/j75689/Tmaster/pkg/database/model"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"github.com/j75689/Tmaster/pkg/lock"
	"github.com/j75689/Tmaster/pkg/message"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/j75689/Tmaster/pkg/opentracer"
	"github.com/j75689/Tmaster/pkg/utils/collection"
	"github.com/j75689/Tmaster/pkg/utils/gzip"
	"github.com/j75689/Tmaster/pkg/utils/parser"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

const (
	_WorkerName = "Scheduler"
)

func NewWorker(
	config config.Config,
	mq mq.MQ,
	lock lock.Locker,
	logger zerolog.Logger,
	lockTimeout time.Duration,
	tracer *opentracer.ServiceTracer,
) *ScheduleWorker {
	return &ScheduleWorker{
		config:      config,
		mq:          mq,
		lock:        lock,
		logger:      logger,
		lockTimeout: lockTimeout,
		tracer:      tracer,
	}
}

type ScheduleWorker struct {
	config      config.Config
	mq          mq.MQ
	lock        lock.Locker
	logger      zerolog.Logger
	lockTimeout time.Duration
	tracer      *opentracer.ServiceTracer
}

func (worker *ScheduleWorker) Process(taskOutput *message.TaskOutput) (*message.TaskInput, error) {
	locked, _ := worker.lock.LockWithAutoDelay(taskOutput.Context.Context, taskOutput.Context.Job.JobID)
	if !locked {
		return nil, fmt.Errorf("duplicate lock for %s", taskOutput.Context.Job.JobID)
	}
	defer worker.lock.UnLock(taskOutput.Context.Context, taskOutput.Context.Job.JobID)

	if worker.config.OpenTracing.Enable {
		var (
			traceRecord *opentracer.TraceRecord
			terr        error
			tags        = map[string]interface{}{"job_id": taskOutput.Context.Job.JobID}
		)
		if taskOutput.Context.TraceCarrier != nil {
			traceRecord, terr = worker.tracer.TraceServer(
				_WorkerName,
				taskOutput.Context.TraceCarrier,
				tags,
			)
		} else {
			traceRecord, terr = worker.tracer.TraceClient(
				_WorkerName,
				tags,
			)
		}
		if terr != nil {
			worker.logger.Err(terr).Msg("create trace error")
		}
		defer traceRecord.Finish()
		taskOutput.Context.TraceCarrier = traceRecord.Carrier()
	}

	var (
		updateJob         = &dbmodel.Job{ID: taskOutput.Context.Job.ID, JobStatus: &model.JobStatus{JobID: taskOutput.Context.Job.JobID}}
		interval          = time.Duration(0)
		retryCount        = 0
		nextTask          *model.Task
		currentTask       = taskOutput.Task
		endTag            = true
		taskInput         *message.TaskInput
		cause             = model.CauseExecute
		consistent        = false
		consistentNums    = taskOutput.Context.Execution.ConsistentNums
		maxConsistentNums = taskOutput.Context.Execution.MaxConsistentNums
		timeout           = taskOutput.Context.Execution.Timeout
		maxTaskExecution  = taskOutput.Context.Execution.MaxTaskExecution
		taskExecution     = taskOutput.Context.Execution.TaskExecution + 1
	)
	defer time.Sleep(interval)

	if taskOutput.Error == nil {
		// task success
		if !reflect.DeepEqual(currentTask.End, &endTag) {
			nextTaskName := ""
			if currentTask.Next != nil {
				nextTaskName = *currentTask.Next
			}
			nextTask = taskOutput.Context.Tasks[nextTaskName]
			updateJob.JobStatus.Status = model.StatusWorking
		} else {
			// update job status
			updateJob.JobStatus.Status = model.StatusSuccess
		}
		updateJob.JobStatus.Timestamp = time.Now()
	} else {
		// task error
		retryCount = taskOutput.Context.State.RetryCount
		errorCode := taskOutput.ErrorCode
		retryMaxAttempts := 0
		if currentTask.Retry != nil && currentTask.Retry.MaxAttempts != nil {
			maxAttemptstr := []byte(*currentTask.Retry.MaxAttempts)
			maxAttemptstr, err := parser.ReplaceVariables(maxAttemptstr, taskOutput.OutputValue)
			if err != nil {
				return nil, fmt.Errorf("replace variable to max attempts error [%v]", err)
			}
			maxAttemptstr, err = parser.ReplaceSystemVariables(maxAttemptstr, taskOutput.OutputValue)
			if err != nil {
				return nil, fmt.Errorf("replace system variable to max attempts error [%v]", err)
			}
			maxAttempts, err := strconv.Atoi(string(maxAttemptstr))
			if err != nil {
				return nil, fmt.Errorf("convert max attempts error [%v]", err)
			}
			retryMaxAttempts = maxAttempts
		}
		if currentTask.Retry != nil &&
			collection.ContainsError(errorCode, currentTask.Retry.ErrorOn) &&
			(retryMaxAttempts > retryCount) &&
			!collection.ContainsErrorMessage(*taskOutput.Error, currentTask.Retry.ExcludeErrorMessage) {
			if sec := currentTask.Retry.Interval; sec != nil {
				interval = time.Second * time.Duration(*sec)
			}
			retryCount++
			cause = model.CauseRetry
			nextTask = &currentTask
		} else if currentTask.Catch != nil &&
			collection.ContainsError(errorCode, currentTask.Catch.ErrorOn) &&
			!collection.ContainsErrorMessage(*taskOutput.Error, currentTask.Catch.ExcludeErrorMessage) {
			retryCount = 0
			cause = model.CauseCatch
			nextTask = taskOutput.Context.Tasks[currentTask.Catch.Next]
		} else {
			retryCount = 0
			updateJob.JobStatus.Status = model.StatusFailed
		}
	}

	hasNext := false
	if nextTask != nil {
		hasNext = true
		now := time.Now()
		id := xid.New().String()

		if maxConsistentNums > consistentNums {
			consistent = true
			consistentNums++
		} else {
			consistentNums = 0
		}

		if timeout != nil && timeout.Before(time.Now()) {
			hasNext = false
			updateJob.JobStatus.Status = model.StatusTimeout

		}

		if taskExecution > maxTaskExecution {
			hasNext = false
			updateJob.JobStatus.Status = model.StatusOverload
		}

		taskInput = &message.TaskInput{
			Context: message.Context{
				TraceCarrier: taskOutput.Context.TraceCarrier,
				Execution: message.Execution{
					ID:                id,
					Cause:             cause,
					CauseError:        taskOutput.Error,
					CauseErrorCode:    taskOutput.ErrorCode,
					ConsistentNums:    consistentNums,
					MaxConsistentNums: maxConsistentNums,
					Timeout:           timeout,
					MaxTaskExecution:  maxTaskExecution,
					TaskExecution:     taskExecution,
				},
				State: message.State{
					EnteredTime: now,
					Name:        nextTask.Name,
					RetryCount:  retryCount,
				},
				Job:   taskOutput.Context.Job,
				Tasks: taskOutput.Context.Tasks,
			},
			Task:       *nextTask,
			From:       taskOutput.TaskXID,
			TaskXID:    id,
			InputValue: taskOutput.OutputValue,
			Consistent: consistent,
		}
	}

	// determine the sequence of update order
	updateJob.JobStatus.Timestamp = time.Now()
	// update job & task
	go func() {
		err := worker.UpdateJob(updateJob)
		if err != nil {
			worker.logger.Err(err).Str("job_id", taskOutput.Context.Job.JobID).Msg("save job error")
		}

		err = worker.CreateTaskStatus(taskOutput)
		if err != nil {
			worker.logger.Err(err).Str("task_id", taskOutput.TaskXID).Msg("save task error")
		}
	}()

	if !hasNext {
		return nil, nil
	}

	return taskInput, nil
}

func (worker *ScheduleWorker) UpdateJob(
	job *dbmodel.Job,
) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	data, err = gzip.Compress(data)
	if err != nil {
		return err
	}
	return worker.mq.Publish(
		time.Now().UnixNano(),
		worker.config.TaskScheduler.JobDBHelper.ProjectID,
		worker.config.TaskScheduler.JobDBHelper.Topic,
		data,
	)
}

func (worker *ScheduleWorker) CreateTaskStatus(
	taskOutput *message.TaskOutput,
) error {
	now := time.Now()
	createdTime := taskOutput.Context.Execution.StartTime
	completeTime := &taskOutput.Context.Execution.EndTime
	cancelTime := &taskOutput.Context.Execution.EndTime
	if taskOutput.Context.State.Status == model.StatusTimeout {
		completeTime = nil
	} else {
		cancelTime = nil
	}
	inputValue, ok := taskOutput.InputValue.(map[string]interface{})
	if !ok {
		inputValue = make(map[string]interface{})
		inputValue["input"] = taskOutput.InputValue
	}
	outputValue, ok := taskOutput.OutputValue.(map[string]interface{})
	if !ok {
		outputValue = make(map[string]interface{})
		outputValue["output"] = taskOutput.OutputValue
	}

	dbTask := dbmodel.Task{
		Name:    taskOutput.Task.Name,
		JobID:   taskOutput.Context.Job.ID,
		Created: createdTime,
		Updated: now,
		TaskHistory: &model.TaskHistory{
			From:        taskOutput.From,
			Cause:       taskOutput.Context.Execution.Cause,
			TaskID:      taskOutput.TaskXID,
			Status:      taskOutput.Context.State.Status,
			ExecutedAt:  &taskOutput.Context.Execution.StartTime,
			CancelledAt: cancelTime,
			CompletedAt: completeTime,
			RetryCount:  &taskOutput.Context.State.RetryCount,
			Input:       inputValue,
			Output:      outputValue,
		},
	}
	if taskOutput.Error != nil {
		dbTask.ErrorMessage = *taskOutput.Error
		dbTask.ErrorCode = *taskOutput.ErrorCode
	}

	data, err := json.Marshal(dbTask)
	if err != nil {
		return err
	}
	data, err = gzip.Compress(data)
	if err != nil {
		return err
	}
	return worker.mq.Publish(
		time.Now().UnixNano(),
		worker.config.TaskScheduler.TaskDBHelper.ProjectID,
		worker.config.TaskScheduler.TaskDBHelper.Topic,
		data,
	)
}
