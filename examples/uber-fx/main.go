package main

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	createApp().Run()
}

func createApp() *fx.App {
	return fx.New(
		provideLogger(),

		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger}
		}),

		provideControllers(),
		provideServerModule(),
	)
}

func provideControllers() fx.Option {
	return fx.Provide(
		NewTestHandler,
	)
}

func provideServerModule() fx.Option {
	return fx.Options(
		provideServer(),
		invokeServer(),
	)
}
