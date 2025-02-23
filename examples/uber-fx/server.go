package main

import (
	"github.com/gopybara/httpbara"
	"go.uber.org/fx"
)

type NewExampleServerIn struct {
	fx.In

	Handlers []*httpbara.Handler `group:"handlers"`
	Log      httpbara.Logger
}

func NewExampleServer(in NewExampleServerIn) (httpbara.Engine, error) {
	logMw, err := httpbara.NewAccessLogMiddleware(in.Log)
	if err != nil {
		return nil, err
	}

	return httpbara.New(in.Handlers,
		httpbara.WithLogger(
			in.Log,
		),
		httpbara.WithRootMiddlewares(logMw),
	)
}

func provideServer() fx.Option {
	return fx.Provide(
		NewExampleServer,
	)
}

func invokeServer() fx.Option {
	return fx.Invoke(
		func(e httpbara.Engine) {
			go e.Run(":8080")
		},
	)
}
