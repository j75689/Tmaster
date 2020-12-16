package resolver

import (
	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/j75689/Tmaster/pkg/opentracer"
	"xorm.io/xorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	config config.Config
	mq     mq.MQ
	db     *xorm.Engine
	tracer *opentracer.ServiceTracer
	logger zerolog.Logger
}

func NewResolver(config config.Config, mq mq.MQ, db *xorm.Engine, tracer *opentracer.ServiceTracer, logger zerolog.Logger) *Resolver {
	return &Resolver{
		config: config,
		mq:     mq,
		db:     db,
		tracer: tracer,
		logger: logger,
	}
}
