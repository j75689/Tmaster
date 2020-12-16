package supported

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/opentracer"
	"github.com/j75689/Tmaster/pkg/opentracer/jaeger"
	"github.com/j75689/Tmaster/pkg/opentracer/zipkin"
)

type OpenTracingDriver string

const (
	Jaeger OpenTracingDriver = "jaeger"
	Zipkin OpenTracingDriver = "zipkin"
)

var supported = map[OpenTracingDriver]func(
	config config.OpenTracingConfig,
	logger zerolog.Logger,
) (opentracer.OpenTracer, error){
	Jaeger: jaeger.NewJaegerTracer,
	Zipkin: zipkin.NewZipKinTracer,
}

func NewOpenTracer(config config.OpenTracingConfig, logger zerolog.Logger) (opentracer.OpenTracer, error) {
	if f, ok := supported[OpenTracingDriver(config.Driver)]; ok {
		return f(config, logger)
	}
	return nil, fmt.Errorf("unsupported opentracing driver [%s]", config.Driver)
}
