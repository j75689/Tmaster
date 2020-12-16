package wireset

import (
	"github.com/google/wire"
	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/opentracer"
	"github.com/j75689/Tmaster/pkg/opentracer/supported"
)

var TracerSet = wire.NewSet(
	InitializeServiceTracer,
	InitializeOpenTracer,
)

func InitializeOpenTracer(config config.Config, logger zerolog.Logger) (opentracer.OpenTracer, error) {
	if config.OpenTracing.Enable {
		return supported.NewOpenTracer(config.OpenTracing, logger)
	}
	return nil, nil
}

func InitializeServiceTracer(config config.Config, tracer opentracer.OpenTracer) *opentracer.ServiceTracer {
	return opentracer.NewServiceTracer(config.OpenTracing, tracer)
}
