package jaeger

import (
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog"
	jaeger "github.com/uber/jaeger-client-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/opentracer"
)

var _ opentracer.OpenTracer = (*JaegerTracer)(nil)

type JaegerTracer struct {
	tracer opentracing.Tracer
	closer io.Closer
}

func (tracer *JaegerTracer) Tracer() opentracing.Tracer {
	return tracer.tracer
}

func (tracer *JaegerTracer) Close() error {
	return tracer.closer.Close()
}

func NewJaegerTracer(config config.OpenTracingConfig, logger zerolog.Logger) (opentracer.OpenTracer, error) {
	cfg := &jaegerConfig.Configuration{
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegerConfig.ReporterConfig{
			LocalAgentHostPort: config.LocalReporter,
			CollectorEndpoint:  config.RemoteReporter,
			LogSpans:           true,
		},
	}
	jLogger := &wrapLogger{logger: logger}
	jMetricsFactory := metrics.NullFactory
	jNullMetrics := jaeger.NewNullMetrics()
	jHeadersConfig := &jaeger.HeadersConfig{
		TraceContextHeaderName: "trace-id",
	}
	tracer, closer, err := cfg.New(
		config.ServiceName,
		jaegerConfig.Logger(jLogger),
		jaegerConfig.Metrics(jMetricsFactory),
		jaegerConfig.Injector(opentracing.TextMap, jaeger.NewTextMapPropagator(jHeadersConfig, *jNullMetrics)),
		jaegerConfig.Extractor(opentracing.TextMap, jaeger.NewTextMapPropagator(jHeadersConfig, *jNullMetrics)),
	)
	if err != nil {
		return nil, err
	}

	return &JaegerTracer{
		tracer: tracer,
		closer: closer,
	}, nil
}

type wrapLogger struct {
	logger zerolog.Logger
}

func (l *wrapLogger) Error(msg string) {
	l.logger.Error().Msg("ERROR: " + msg)
}

func (l *wrapLogger) Infof(msg string, args ...interface{}) {
	l.logger.Info().Msgf(msg, args...)
}
