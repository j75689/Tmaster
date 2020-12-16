package scheduler

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
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"xorm.io/xorm"
)

type Application struct {
	db     *xorm.Engine
	config config.Config
	logger zerolog.Logger
	worker *scheduler.ScheduleWorker
	mq     mq.MQ
	tracer *opentracer.ServiceTracer
}

func (application Application) Start() error {
	application.logger.Info().Msg("task scheduler program is working")
	g := errgroup.Group{}
	cfg := application.config.TaskScheduler
	g.Go(func() error {
		if err := application.mq.Subscribe(cfg.TaskOutput.ProjectID, cfg.TaskOutput.SubscribeID, func(ctx context.Context, data []byte) error {
			ft := time.Now()

			t := time.Now()
			data, err := gzip.Decompress(data)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("decompress input message")

			t = time.Now()
			application.logger.Trace().Bytes("data", data).Msg("recived message")
			var taskOutput message.TaskOutput
			err = json.Unmarshal(data, &taskOutput)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("unmarshal input message")

			t = time.Now()
			taskInput, err := application.worker.Process(&taskOutput)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("process input message")
			if taskInput == nil {
				return nil
			}

			t = time.Now()
			outputData, err := json.Marshal(taskInput)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("marshal output message")

			t = time.Now()
			data, err = gzip.Compress(outputData)
			if err != nil {
				return err
			}

			defer func() {
				application.logger.Debug().Str("task_id", taskInput.TaskXID).Dur("duration", time.Since(ft)).Msg("completed message")
			}()
			return application.mq.Publish(time.Now().UnixNano(), cfg.TaskInput.ProjectID, cfg.TaskInput.Topic, data)
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
	err := application.db.Close()
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
	db *xorm.Engine,
	worker *scheduler.ScheduleWorker,
	logger zerolog.Logger,
	tracer *opentracer.ServiceTracer,
) Application {
	return Application{
		db:     db,
		config: config,
		logger: logger,
		worker: worker,
		mq:     mq,
		tracer: tracer,
	}
}
