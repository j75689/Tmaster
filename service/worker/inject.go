package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/message"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/j75689/Tmaster/pkg/opentracer"
	"github.com/j75689/Tmaster/pkg/utils/gzip"
	"github.com/j75689/Tmaster/pkg/worker/scheduler"
	"github.com/j75689/Tmaster/pkg/worker/task"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type Application struct {
	config    config.Config
	logger    zerolog.Logger
	worker    *task.TaskWorker
	scheduler *scheduler.ScheduleWorker
	db        *gorm.DB
	mq        mq.MQ
	tracer    *opentracer.ServiceTracer
}

func (application Application) Start() error {
	application.logger.Info().Msg("task worker program is working")
	g := errgroup.Group{}
	cfg := application.config.TaskWorker
	g.Go(func() error {
		if err := application.mq.Subscribe(cfg.TaskInput.ProjectID, cfg.TaskInput.SubscribeID, func(ctx context.Context, data []byte) error {
			ft := time.Now()
			t := time.Now()
			data, err := gzip.Decompress(data)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("decompress input message")

			t = time.Now()
			application.logger.Trace().Bytes("data", data).Msg("received message")
			var taskInput message.TaskInput
			err = json.Unmarshal(data, &taskInput)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("unmarshal input message")

			var (
				taskOutput *message.TaskOutput
				consistent = true
			)

			for consistent {
				t = time.Now()
				taskOutput, err = application.worker.Process(&taskInput)
				if err != nil {
					return err
				}
				application.logger.Debug().Dur("duration", time.Since(t)).Msg("process task input message")
				consistent = taskInput.Consistent
				if taskInput.Consistent {
					t = time.Now()
					nextTaskInput, err := application.scheduler.Process(taskOutput)
					if err != nil {
						return err
					}
					application.logger.Debug().Dur("duration", time.Since(t)).Msg("process scheduler input message")
					if nextTaskInput == nil {
						return nil
					}
					taskInput = *nextTaskInput
				}
			}

			t = time.Now()
			outputData, err := json.Marshal(taskOutput)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("marshal output message")

			t = time.Now()
			data, err = gzip.Compress(outputData)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("compress output message")

			defer func() {
				application.logger.Debug().Str("task_id", taskInput.TaskXID).Dur("duration", time.Since(ft)).Msg("completed message")
			}()

			return application.mq.Publish(time.Now().UnixNano(), cfg.TaskOutput.ProjectID, cfg.TaskOutput.Topic, data)
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
	worker *task.TaskWorker,
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
		scheduler: scheduler,
		tracer:    tracer,
	}
}
