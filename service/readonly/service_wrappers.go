// Code generated by otelwrap; DO NOT EDIT.
// github.com/QuangTung97/otelwrap

package readonly

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

// IServiceWrapper wraps OpenTelemetry's span
type IServiceWrapper struct {
	IService
	tracer trace.Tracer
	prefix string
}

// NewIServiceWrapper creates a wrapper
func NewIServiceWrapper(wrapped IService, tracer trace.Tracer, prefix string) *IServiceWrapper {
	return &IServiceWrapper{
		IService: wrapped,
		tracer:   tracer,
		prefix:   prefix,
	}
}

// Check ...
func (w *IServiceWrapper) Check(ctx context.Context, inputs []Input) (a []Output) {
	ctx, span := w.tracer.Start(ctx, w.prefix+"Check")
	defer span.End()

	a = w.IService.Check(ctx, inputs)

	return a
}
