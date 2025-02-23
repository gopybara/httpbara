package httpbaratelemetry

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type otelMiddlewareDescriber struct {
	InjectTrace httpbara.Middleware `middleware:"otelInjector"`
}

type otelMiddleware struct {
	otelMiddlewareDescriber

	tp TelemetryProvider
}

func NewOtelMiddleware(tp TelemetryProvider) (*httpbara.Handler, error) {
	omi := otelMiddleware{
		tp: tp,
	}

	return httpbara.AsHandler(&omi)
}

func (omi *otelMiddleware) InjectTrace(ctx *gin.Context) {
	spanName := ctx.Request.Method + " " + ctx.FullPath()
	var traceCtx context.Context
	var span trace.Span

	if ctx.GetHeader("traceparent") != "" {
		ctx.Request = ctx.Request.WithContext(omi.tp.propagator().Extract(ctx, propagation.HeaderCarrier(ctx.Request.Header)))

		span = trace.SpanFromContext(ctx.Request.Context())
		traceCtx = trace.ContextWithSpan(ctx.Request.Context(), span)
	} else {
		traceCtx, span = omi.tp.NewSpan(ctx.Request.Context(), spanName)

		ctx.Request.Header.Set("Traceparent", omi.tp.createTraceparent(traceCtx))

		omi.tp.propagator().Inject(traceCtx, propagation.HeaderCarrier(ctx.Request.Header))

		ctx.Request = ctx.Request.WithContext(traceCtx)
	}

	defer span.End()

	ctx.Next()
}
