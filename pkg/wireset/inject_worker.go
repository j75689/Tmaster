package wireset

import (
	"github.com/google/wire"
	"github.com/j75689/Tmaster/pkg/config"
	choiceExecutor "github.com/j75689/Tmaster/pkg/executor/choice"
	parallelExecutor "github.com/j75689/Tmaster/pkg/executor/parallel"
	passExecutor "github.com/j75689/Tmaster/pkg/executor/pass"
	taskExecutor "github.com/j75689/Tmaster/pkg/executor/task"
	waitExecutor "github.com/j75689/Tmaster/pkg/executor/wait"
	"github.com/j75689/Tmaster/pkg/lock"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/j75689/Tmaster/pkg/opentracer"
	"github.com/j75689/Tmaster/pkg/worker/dbhelper"
	"github.com/j75689/Tmaster/pkg/worker/initializer"
	"github.com/j75689/Tmaster/pkg/worker/scheduler"
	"github.com/j75689/Tmaster/pkg/worker/task"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var WorkerSet = wire.NewSet(
	InitializeWorker,
	InitializeTaskWorker,
	InitializeScheduleWorker,
	InitializeDBHelperWorker,
)

func InitializeWorker(
	config config.Config,
	db *gorm.DB,
	lock lock.Locker,
	logger zerolog.Logger,
	tracer *opentracer.ServiceTracer,
) *initializer.InitializeWorker {
	return initializer.NewWorker(config, db, lock, logger, config.Redis.LockTimeout, tracer)
}

func InitializeTaskWorker(
	taskExecutor *taskExecutor.TaskExecutor,
	parallelExecutor *parallelExecutor.ParallelExecutor,
	waitExecutor *waitExecutor.WaitExecutor,
	passExecutor *passExecutor.PassExecutor,
	choiceExecutor *choiceExecutor.ChoiceExecutor,
	config config.Config,
	lock lock.Locker,
	logger zerolog.Logger,
	tracer *opentracer.ServiceTracer,
) *task.TaskWorker {
	return task.NewWorker(
		config,
		taskExecutor, parallelExecutor, waitExecutor, passExecutor, choiceExecutor,
		lock, logger, config.Redis.LockTimeout, tracer,
	)
}

func InitializeScheduleWorker(
	config config.Config,
	mq mq.MQ,
	lock lock.Locker,
	logger zerolog.Logger,
	tracer *opentracer.ServiceTracer,
) *scheduler.ScheduleWorker {
	return scheduler.NewWorker(config, mq, lock, logger, config.Redis.LockTimeout, tracer)
}

func InitializeDBHelperWorker(
	config config.Config,
	db *gorm.DB,
) *dbhelper.DBHelperWorker {
	return dbhelper.NewWorker(config, db)
}
