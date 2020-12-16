package zipkin

import (
	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/opentracer"
)

var _ opentracer.OpenTracer = (*ZipKinTracer)(nil)

func NewZipKinTracer(config config.OpenTracingConfig, logger zerolog.Logger) (opentracer.OpenTracer, error) {
	reporter := zipkinhttp.NewReporter(config.RemoteReporter)
	endpoint, err := zipkin.NewEndpoint(config.ServiceName, config.LocalReporter)
	if err != nil {
		return nil, err
	}
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		return nil, err
	}

	// use zipkin-go-opentracing to wrap our tracer
	tracer := zipkinot.Wrap(nativeTracer)

	return &ZipKinTracer{
		reporter: reporter,
		tracer:   tracer,
	}, nil
}

type ZipKinTracer struct {
	reporter reporter.Reporter
	tracer   opentracing.Tracer
}

func (tracer *ZipKinTracer) Tracer() opentracing.Tracer {
	return tracer.tracer
}

func (tracer *ZipKinTracer) Close() error {
	return tracer.reporter.Close()
}
