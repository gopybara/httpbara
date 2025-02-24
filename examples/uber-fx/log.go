package main

import (
	"github.com/gopybara/httpbara"
	"github.com/gopybara/httpbara/pkg/httpbarazap"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func newZap() (*zap.Logger, error) {
	return zap.NewProduction()
}

func newHttpbaraLogger(log *zap.Logger) httpbara.Logger {
	return httpbarazap.New(
		log,
	)
}

func provideLogger() fx.Option {
	return fx.Provide(
		newZap,
		newHttpbaraLogger,
		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger}
		}),
	)
}
