package initializer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/config"
	dbmodel "github.com/j75689/Tmaster/pkg/database/model"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"github.com/j75689/Tmaster/pkg/lock"
	"github.com/j75689/Tmaster/pkg/message"
	"github.com/j75689/Tmaster/pkg/opentracer"
	"github.com/j75689/Tmaster/pkg/utils/parser"
	"xorm.io/xorm"
)

const (
	_WorkerName = "Initializer"
)

func NewWorker(
	config config.Config,
	db *xorm.Engine,
	lock lock.Locker,
	logger zerolog.Logger,
	lockTimeout time.Duration,
	tracer *opentracer.ServiceTracer,
) *InitializeWorker {
	return &InitializeWorker{
		config:      config,
		db:          db,
		lock:        lock,
		logger:      logger,
		lockTimeout: lockTimeout,
		tracer:      tracer,
	}
}

type InitializeWorker struct {
	config      config.Config
	db          *xorm.Engine
	lock        lock.Locker
	logger      zerolog.Logger
	lockTimeout time.Duration
	tracer      *opentracer.ServiceTracer
}

func (worker *InitializeWorker) Process(initJob *message.InitJob) (*message.TaskInput, error) {
	locked, _ := worker.lock.LockWithAutoDelay(initJob.Context.Context, initJob.JobID)
	if !locked {
		return nil, fmt.Errorf("duplicate lock for %s", initJob.JobID)
	}
	defer worker.lock.UnLock(initJob.Context.Context, initJob.JobID)

	if worker.config.OpenTracing.Enable {
		var (
			traceRecord *opentracer.TraceRecord
			terr        error
			tags        = map[string]interface{}{"job_id": initJob.JobID}
		)
		if initJob.Context.TraceCarrier != nil {
			traceRecord, terr = worker.tracer.TraceServer(
				_WorkerName,
				initJob.Context.TraceCarrier,
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
		initJob.Context.TraceCarrier = traceRecord.Carrier()
	}

	// insert job
	now := time.Now()
	job := &dbmodel.Job{
		ParentID: initJob.ParentID,
		Created:  now,
		JobStatus: &model.JobStatus{
			JobID:     initJob.JobID,
			Status:    model.StatusPending,
			Timestamp: now,
		},
	}
	if _, err := worker.db.InsertOne(job); err != nil {
		return nil, err
	}

	var taskInput *message.TaskInput
	id := xid.New().String()

	inputValue := make(map[string]interface{})
	if initJob.Parameters != nil && *initJob.Parameters != "" {
		if err := json.Unmarshal([]byte(*initJob.Parameters), &inputValue); err != nil {
			return nil, err
		}
	}
	maxConsistentNums := 0
	consistent := false
	if initJob.Job.ConsistentTaskNums != nil {
		maxConsistentNums = *initJob.Job.ConsistentTaskNums
		consistent = maxConsistentNums > 0
	}

	taskMap := parser.TaskArrayToMap(initJob.Tasks)
	task := taskMap[initJob.StartAt]
	if task != nil {
		taskInput = &message.TaskInput{
			Context: message.Context{
				TraceCarrier: initJob.Context.TraceCarrier,
				Execution: message.Execution{
					ID:                id,
					Cause:             model.CauseExecute,
					MaxConsistentNums: maxConsistentNums,
				},
				State: message.State{
					EnteredTime: now,
					Name:        task.Name,
				},
				Job: message.Job{
					ID:    job.ID,
					JobID: initJob.JobID,
				},
				Tasks: taskMap,
			},
			Task:       *task,
			TaskXID:    id,
			InputValue: inputValue,
			Consistent: consistent,
		}
	} else {
		job.JobStatus.Status = model.StatusSuccess
		worker.db.Update(job)
	}

	return taskInput, nil
}
