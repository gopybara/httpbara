package httpbaratelemetry

import (
	"context"
	"fmt"
	"github.com/gopybara/httpbara"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrTracerProviderNotSet = fmt.Errorf("tracer provider not set")
)

type TelemetryProvider interface {
	LogWithContext(ctx context.Context) httpbara.Logger
	LogWithoutContext() httpbara.Logger

	NewSpan(ctx context.Context, name string, attributes ...attribute.KeyValue) (context.Context, trace.Span)
	CurrentSpan(ctx context.Context) trace.Span

	createTraceparent(ctx context.Context) string
	propagator() propagation.TextMapPropagator
	Provider() *sdktrace.TracerProvider
}

type loggerWithContext struct {
	propagator    propagation.TextMapPropagator
	tp            *sdktrace.TracerProvider
	ctx           context.Context
	l             httpbara.Logger
	telemetryKeys *TelemetryKeys
}

func (lwc *loggerWithContext) addSpanToFields(fields *[]any) {
	if span := trace.SpanFromContext(lwc.ctx); span != nil {
		*fields = append(*fields, lwc.telemetryKeys.TraceID, span.SpanContext().TraceID().String(), lwc.telemetryKeys.SpanID, span.SpanContext().SpanID().String())
	}
}

func (lwc *loggerWithContext) Info(msg string, fields ...any) {
	lwc.addSpanToFields(&fields)

	lwc.l.Info(msg, fields...)
}

func (lwc *loggerWithContext) Error(msg string, fields ...any) {
	lwc.addSpanToFields(&fields)

	lwc.l.Error(msg, fields...)
}

func (lwc *loggerWithContext) Debug(msg string, fields ...any) {
	lwc.addSpanToFields(&fields)

	lwc.l.Debug(msg, fields...)
}

func (lwc *loggerWithContext) Warn(msg string, fields ...any) {
	lwc.addSpanToFields(&fields)

	lwc.l.Warn(msg, fields...)
}

func (lwc *loggerWithContext) Panic(msg string, fields ...any) {
	lwc.addSpanToFields(&fields)

	lwc.l.Panic(msg, fields...)
}

type telemetryOpts struct {
	log httpbara.Logger

	tracerName    string `default:"httpbara"`
	traceProvider *sdktrace.TracerProvider
	// Can be empty
	// If empty propagation.TraceContext will be used by default
	propagator propagation.TextMapPropagator

	telemetryKeys *TelemetryKeys
}

type TelemetryKeys struct {
	TraceID string `default:"trace_id"`
	SpanID  string `default:"span_id"`
}

type TelemetryOpt func(*telemetryOpts)

func WithTracerName(name string) TelemetryOpt {
	return func(opts *telemetryOpts) {
		opts.tracerName = name
	}
}

func WithTelemetryLogger(log httpbara.Logger) TelemetryOpt {
	return func(opts *telemetryOpts) {
		opts.log = log
	}
}

func WithTelemetryKeys(keys *TelemetryKeys) TelemetryOpt {
	return func(opts *telemetryOpts) {
		opts.telemetryKeys = keys
	}
}

func WithTraceProvider(tp *sdktrace.TracerProvider) TelemetryOpt {
	return func(opts *telemetryOpts) {
		opts.traceProvider = tp
	}
}

type providerImpl struct {
	opts telemetryOpts
}

func (pi *providerImpl) Provider() *sdktrace.TracerProvider {
	return pi.opts.traceProvider
}

func (pi *providerImpl) LogWithContext(ctx context.Context) httpbara.Logger {
	return &loggerWithContext{
		ctx:           ctx,
		propagator:    pi.opts.propagator,
		tp:            pi.opts.traceProvider,
		l:             pi.opts.log,
		telemetryKeys: pi.opts.telemetryKeys,
	}
}

func (pi *providerImpl) LogWithoutContext() httpbara.Logger {
	return pi.opts.log
}

func (pi *providerImpl) NewSpan(ctx context.Context, name string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	return pi.opts.traceProvider.Tracer(pi.opts.tracerName).Start(ctx, name, trace.WithAttributes(attributes...))
}

func (pi *providerImpl) CurrentSpan(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func (pi *providerImpl) createTraceparent(ctx context.Context) string {
	sc := trace.SpanFromContext(ctx).SpanContext()

	return fmt.Sprintf("00-%s-%s-%s",
		sc.TraceID().String(),
		sc.SpanID().String(),
		sc.TraceFlags().String(),
	)
}

func (pi *providerImpl) propagator() propagation.TextMapPropagator {
	return pi.opts.propagator
}

func NewProvider(opts ...TelemetryOpt) (TelemetryProvider, error) {
	to := telemetryOpts{
		telemetryKeys: &TelemetryKeys{
			TraceID: "trace_id",
			SpanID:  "span_id",
		},
		propagator: propagation.TraceContext{},
		tracerName: "httpbara",
	}

	for _, opt := range opts {
		opt(&to)
	}

	if to.log == nil {
		to.log = httpbara.NewFmtLogger()
	}

	if to.traceProvider == nil {
		return nil, ErrTracerProviderNotSet
	}

	return &providerImpl{
		opts: to,
	}, nil
}
