// Code generated by otelwrap; DO NOT EDIT.
// github.com/QuangTung97/otelwrap

package repository

import (
	"context"
	"github.com/QuangTung97/promo-readonly/model"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// BlacklistWrapper wraps OpenTelemetry's span
type BlacklistWrapper struct {
	Blacklist
	tracer trace.Tracer
	prefix string
}

// NewBlacklistWrapper creates a wrapper
func NewBlacklistWrapper(wrapped Blacklist, tracer trace.Tracer, prefix string) *BlacklistWrapper {
	return &BlacklistWrapper{
		Blacklist: wrapped,
		tracer:    tracer,
		prefix:    prefix,
	}
}

// GetConfig ...
func (w *BlacklistWrapper) GetConfig(ctx context.Context) (a model.BlacklistConfig, err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix+"GetConfig")
	defer span.End()

	a, err = w.Blacklist.GetConfig(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return a, err
}

// UpsertConfig ...
func (w *BlacklistWrapper) UpsertConfig(ctx context.Context, config model.BlacklistConfig) (err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix+"UpsertConfig")
	defer span.End()

	err = w.Blacklist.UpsertConfig(ctx, config)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

// GetBlacklistCustomers ...
func (w *BlacklistWrapper) GetBlacklistCustomers(ctx context.Context, keys []BlacklistCustomerKey) (a []model.BlacklistCustomer, err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix+"GetBlacklistCustomers")
	defer span.End()

	a, err = w.Blacklist.GetBlacklistCustomers(ctx, keys)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return a, err
}

// SelectBlacklistCustomers ...
func (w *BlacklistWrapper) SelectBlacklistCustomers(ctx context.Context, ranges []HashRange) (a []model.BlacklistCustomer, err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix+"SelectBlacklistCustomers")
	defer span.End()

	a, err = w.Blacklist.SelectBlacklistCustomers(ctx, ranges)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return a, err
}

// UpsertBlacklistCustomers ...
func (w *BlacklistWrapper) UpsertBlacklistCustomers(ctx context.Context, customers []model.BlacklistCustomer) (err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix+"UpsertBlacklistCustomers")
	defer span.End()

	err = w.Blacklist.UpsertBlacklistCustomers(ctx, customers)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

// GetBlacklistMerchants ...
func (w *BlacklistWrapper) GetBlacklistMerchants(ctx context.Context, keys []BlacklistMerchantKey) (a []model.BlacklistMerchant, err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix+"GetBlacklistMerchants")
	defer span.End()

	a, err = w.Blacklist.GetBlacklistMerchants(ctx, keys)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return a, err
}

// SelectBlacklistMerchants ...
func (w *BlacklistWrapper) SelectBlacklistMerchants(ctx context.Context, ranges []HashRange) (a []model.BlacklistMerchant, err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix+"SelectBlacklistMerchants")
	defer span.End()

	a, err = w.Blacklist.SelectBlacklistMerchants(ctx, ranges)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return a, err
}

// UpsertBlacklistMerchants ...
func (w *BlacklistWrapper) UpsertBlacklistMerchants(ctx context.Context, merchants []model.BlacklistMerchant) (err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix+"UpsertBlacklistMerchants")
	defer span.End()

	err = w.Blacklist.UpsertBlacklistMerchants(ctx, merchants)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

// GetBlacklistTerminals ...
func (w *BlacklistWrapper) GetBlacklistTerminals(ctx context.Context, keys []BlacklistTerminalKey) (a []model.BlacklistTerminal, err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix+"GetBlacklistTerminals")
	defer span.End()

	a, err = w.Blacklist.GetBlacklistTerminals(ctx, keys)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return a, err
}

// UpsertBlacklistTerminals ...
func (w *BlacklistWrapper) UpsertBlacklistTerminals(ctx context.Context, terminals []model.BlacklistTerminal) (err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix+"UpsertBlacklistTerminals")
	defer span.End()

	err = w.Blacklist.UpsertBlacklistTerminals(ctx, terminals)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}