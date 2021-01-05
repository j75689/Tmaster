package graphql

import (
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/graph"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/j75689/Tmaster/pkg/opentracer"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Application struct {
	config     config.Config
	httpServer *graph.HttpServer
	mq         mq.MQ
	db         *gorm.DB
	tracer     *opentracer.ServiceTracer
	logger     zerolog.Logger
}

func (application Application) Start() error {
	return application.httpServer.Start()
}

func (application Application) Shutdown() error {
	application.logger.Info().Msg("stop http server ...")
	if err := application.httpServer.Shutdown(); err != nil {
		return err
	}
	application.logger.Info().Msg("http server stopped")

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

func newApplication(config config.Config, httpServer *graph.HttpServer, mq mq.MQ, db *gorm.DB, tracer *opentracer.ServiceTracer, logger zerolog.Logger) Application {
	return Application{
		config:     config,
		httpServer: httpServer,
		mq:         mq,
		db:         db,
		tracer:     tracer,
		logger:     logger,
	}
}
