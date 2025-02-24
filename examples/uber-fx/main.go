package main

import (
	"go.uber.org/fx"
)

func main() {
	createApp().Run()
}

func createApp() *fx.App {
	return fx.New(
		provideLogger(),

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
