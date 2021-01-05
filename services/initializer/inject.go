package initializer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/message"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/j75689/Tmaster/pkg/opentracer"
	"github.com/j75689/Tmaster/pkg/utils/gzip"
	"github.com/j75689/Tmaster/pkg/worker/initializer"
	"github.com/j75689/Tmaster/pkg/worker/scheduler"
	"github.com/j75689/Tmaster/pkg/worker/task"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type Application struct {
	config    config.Config
	logger    zerolog.Logger
	mq        mq.MQ
	db        *gorm.DB
	worker    *initializer.InitializeWorker
	task      *task.TaskWorker
	scheduler *scheduler.ScheduleWorker
	tracer    *opentracer.ServiceTracer
}

func (application Application) Start() error {
	application.logger.Info().Msg("job initialization program is working")

	g := errgroup.Group{}
	cfg := application.config.JobInitializer
	g.Go(func() error {
		if err := application.mq.Subscribe(cfg.InitJob.ProjectID, cfg.InitJob.SubscribeID, func(ctx context.Context, data []byte) error {
			ft := time.Now()

			t := time.Now()
			data, err := gzip.Decompress(data)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("decompress input message")

			t = time.Now()
			application.logger.Trace().Bytes("data", data).Msg("recived message")
			var initJob message.InitJob
			err = json.Unmarshal(data, &initJob)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("unmarshal input message")

			t = time.Now()
			taskInput, err := application.worker.Process(&initJob)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("process input message")

			if taskInput == nil {
				return nil
			}

			var (
				taskOutput *message.TaskOutput
				consistent = taskInput.Consistent
			)

			for consistent {
				t = time.Now()
				taskOutput, err = application.task.Process(taskInput)
				if err != nil {
					return err
				}
				application.logger.Debug().Dur("duration", time.Since(t)).Msg("process task input message")

				t = time.Now()
				nextTaskInput, err := application.scheduler.Process(taskOutput)
				if err != nil {
					return err
				}
				application.logger.Debug().Dur("duration", time.Since(t)).Msg("process scheduler input message")
				if nextTaskInput == nil {
					return nil
				}
				taskInput = nextTaskInput
				consistent = taskInput.Consistent
			}

			t = time.Now()
			inputData, err := json.Marshal(taskInput)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("marshal output message")

			t = time.Now()
			data, err = gzip.Compress(inputData)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("compress output message")

			defer func() {
				application.logger.Debug().Str("task_id", taskInput.TaskXID).Dur("duration", time.Since(ft)).Msg("completed message")
			}()
			if inputData != nil {
				return application.mq.Publish(time.Now().UnixNano(), cfg.TaskInput.ProjectID, cfg.TaskInput.Topic, data)
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (application Application) Shutdown() error {
	application.logger.Info().Msg("stop mq ...")
	application.mq.Stop()
	application.logger.Info().Msg("mq stopped")

	application.logger.Info().Msg("close db ...")
	db, err := application.db.DB()
	if err != nil {
		return err
	}
	err = db.Close()
	if err != nil {
		return err
	}
	application.logger.Info().Msg("db closed")

	application.logger.Info().Msg("close tracer ...")
	err = application.tracer.Close()
	if err != nil {
		return err
	}
	application.logger.Info().Msg("tracer closed")
	return nil
}

func newApplication(
	config config.Config,
	mq mq.MQ,
	db *gorm.DB,
	worker *initializer.InitializeWorker,
	task *task.TaskWorker,
	scheduler *scheduler.ScheduleWorker,
	logger zerolog.Logger,
	tracer *opentracer.ServiceTracer,
) Application {
	return Application{
		config:    config,
		logger:    logger,
		mq:        mq,
		db:        db,
		worker:    worker,
		task:      task,
		scheduler: scheduler,
		tracer:    tracer,
	}
}
