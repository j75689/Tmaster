package dbhelper

import (
	"context"
	"encoding/json"
	"time"

	"github.com/j75689/Tmaster/pkg/config"
	dbmodel "github.com/j75689/Tmaster/pkg/database/model"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/j75689/Tmaster/pkg/utils/gzip"
	"github.com/j75689/Tmaster/pkg/worker/dbhelper"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type Application struct {
	db     *gorm.DB
	config config.Config
	logger zerolog.Logger
	worker *dbhelper.DBHelperWorker
	mq     mq.MQ
}

func (application Application) Start() error {
	application.logger.Info().Msg("db helper program is working")
	g := errgroup.Group{}
	cfg := application.config.DBHelper
	g.Go(func() error {
		if err := application.mq.Subscribe(cfg.Job.ProjectID, cfg.Job.SubscribeID, func(ctx context.Context, data []byte) error {
			ft := time.Now()

			t := time.Now()
			data, err := gzip.Decompress(data)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("decompress input message")

			t = time.Now()
			application.logger.Trace().Bytes("data", data).Msg("received message")
			var job dbmodel.Job
			err = json.Unmarshal(data, &job)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("unmarshal input message")

			t = time.Now()
			err = application.worker.UpdateJob(&job)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("process update job")
			defer func() {
				application.logger.Debug().Str("job_id", job.JobStatus.JobID).Dur("duration", time.Since(ft)).Msg("completed message")
			}()
			return nil
		}); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		if err := application.mq.Subscribe(cfg.Task.ProjectID, cfg.Task.SubscribeID, func(ctx context.Context, data []byte) error {
			ft := time.Now()

			t := time.Now()
			data, err := gzip.Decompress(data)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("decompress input message")

			t = time.Now()
			application.logger.Trace().Bytes("data", data).Msg("received message")
			var Task dbmodel.Task
			err = json.Unmarshal(data, &Task)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("unmarshal input message")

			t = time.Now()
			err = application.worker.CreateTask(&Task)
			if err != nil {
				return err
			}
			application.logger.Debug().Dur("duration", time.Since(t)).Msg("process create task")
			defer func() {
				application.logger.Debug().Str("task_id", Task.TaskHistory.TaskID).Dur("duration", time.Since(ft)).Msg("completed message")
			}()
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
	return nil
}

func newApplication(
	config config.Config,
	mq mq.MQ,
	db *gorm.DB,
	worker *dbhelper.DBHelperWorker,
	logger zerolog.Logger,
) Application {
	return Application{
		db:     db,
		config: config,
		logger: logger,
		worker: worker,
		mq:     mq,
	}
}
