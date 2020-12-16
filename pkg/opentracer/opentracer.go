package opentracer

import (
	"github.com/opentracing/opentracing-go"
)

type OpenTracer interface {
	Tracer() opentracing.Tracer
	Close() error
}
