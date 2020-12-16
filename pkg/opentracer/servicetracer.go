package opentracer

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/j75689/Tmaster/pkg/config"
)

func NewServiceTracer(config config.OpenTracingConfig, tracer OpenTracer) *ServiceTracer {
	if tracer != nil {
		opentracing.SetGlobalTracer(tracer.Tracer())
	}
	return &ServiceTracer{
		config: config,
		tracer: tracer,
	}
}

type TraceRecord struct {
	span    opentracing.Span
	carrier opentracing.TextMapCarrier
}

func (record *TraceRecord) Finish() {
	if record.span != nil {
		record.span.Finish()
	}
}

func (record *TraceRecord) Carrier() opentracing.TextMapCarrier {
	if record.carrier == nil {
		return opentracing.TextMapCarrier{}
	}
	return record.carrier
}

type ServiceTracer struct {
	config config.OpenTracingConfig
	tracer OpenTracer
}

func (tracer *ServiceTracer) TraceClient(operationName string, tags map[string]interface{}) (*TraceRecord, error) {
	if !tracer.config.Enable {
		return &TraceRecord{}, nil
	}

	opentracer := opentracing.GlobalTracer()
	carrier := opentracing.TextMapCarrier{}
	span := opentracer.StartSpan(operationName)
	err := opentracer.Inject(span.Context(), opentracing.TextMap, carrier)
	if err != nil {
		return &TraceRecord{}, err
	}
	for k, v := range tags {
		span = span.SetTag(k, v)
	}

	return &TraceRecord{span, carrier}, nil
}

func (tracer *ServiceTracer) TraceServer(operationName string, carrier opentracing.TextMapCarrier, tags map[string]interface{}) (*TraceRecord, error) {
	if !tracer.config.Enable {
		return &TraceRecord{}, nil
	}

	opentracer := opentracing.GlobalTracer()
	spanCtx, err := opentracer.Extract(opentracing.TextMap, carrier)
	if err != nil {
		return &TraceRecord{}, err
	}

	span := opentracing.StartSpan(
		operationName,
		ext.RPCServerOption(spanCtx),
	)

	for k, v := range tags {
		span = span.SetTag(k, v)
	}

	return &TraceRecord{span, carrier}, nil
}

func (tracer *ServiceTracer) Close() error {
	if tracer.tracer != nil {
		return tracer.tracer.Close()
	}
	return nil
}
