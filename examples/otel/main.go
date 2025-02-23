package main

import (
	"context"
	"fmt"
	"github.com/gopybara/httpbara"
	"github.com/gopybara/httpbara/pkg/httpbaratelemetry"
)

func main() {
	handler, err := NewTestHandler()
	if err != nil {
		panic(fmt.Errorf("failed to create router: %w", err))
	}

	tp, err := initOtel(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to init otel: %w", err))
	}

	otelProvider, err := httpbaratelemetry.NewProvider(
		httpbaratelemetry.WithTraceProvider(tp),
	)
	if err != nil {
		panic(fmt.Errorf("failed to create otel provider: %w", err))
	}

	mw, err := httpbaratelemetry.NewOtelMiddleware(otelProvider)
	if err != nil {
		panic(fmt.Errorf("failed to create otel middleware: %w", err))
	}

	router, err := httpbara.New(
		[]*httpbara.Handler{handler},
		httpbara.WithRootMiddlewares(mw),
	)
	if err != nil {
		panic(fmt.Errorf("failed to create router: %w", err))
	}

	router.Run(":8080")
}
