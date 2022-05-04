package otellib

import (
	"context"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ctxLoggerKey struct{}
type ctxLoggerValue struct {
	logger *zap.Logger
}

var loggerKey ctxLoggerKey

const (
	traceIDField    = "trace.id"
	spanIDField     = "span.id"
	traceFlagsField = "trace.flags"
)

// SetTraceInfoInterceptor ...
func SetTraceInfoInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		tags := grpc_ctxtags.Extract(ctx)
		sc := trace.SpanContextFromContext(ctx)

		tags.Set(traceIDField, sc.TraceID())
		tags.Set(spanIDField, sc.SpanID())
		tags.Set(traceFlagsField, sc.TraceFlags())

		ctx = context.WithValue(ctx, loggerKey, ctxLoggerValue{logger: logger})
		return handler(ctx, req)
	}
}

// Extract ...
func Extract(ctx context.Context) *zap.Logger {
	val, ok := ctx.Value(loggerKey).(ctxLoggerValue)
	if !ok {
		return zap.NewNop()
	}
	sc := trace.SpanContextFromContext(ctx)
	return val.logger.With(
		zap.String(traceIDField, sc.TraceID().String()),
		zap.String(spanIDField, sc.SpanID().String()),
		zap.String(traceFlagsField, sc.TraceFlags().String()),
	)
}

// WrapError ...
func WrapError(ctx context.Context, err error) {
	Extract(ctx).WithOptions(zap.AddCallerSkip(2)).
		Error("WrapError", zap.Error(err))
}

// ToContext ...
func ToContext(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, ctxLoggerValue{logger: l})
}
