package httpbarafx

import (
	"fmt"
	"github.com/gopybara/httpbara"
	"go.uber.org/fx"
)

type NewExampleServerIn struct {
	fx.In

	Handlers []*httpbara.Handler `group:"handlers"`
	Opts     []httpbara.ParamsCb `group:"httpbaraOpts" optional:"true"`
}

func NewHttpbaraServer(in NewExampleServerIn) (httpbara.Engine, error) {
	return httpbara.New(in.Handlers,
		in.Opts...,
	)
}

func ProvideHttpbaraModule() fx.Option {
	return fx.Options(
		fx.Provide(
			NewHttpbaraServer,
		),
		invokeServer(),
	)
}

type InvokeServerIn struct {
	fx.In

	Engine httpbara.Engine
	Params *HttpbaraRunParams `optional:"true"`
}

func invokeServer() fx.Option {
	return fx.Invoke(
		func(in InvokeServerIn) {
			if in.Params == nil {
				in.Params = &HttpbaraRunParams{Port: 1489}
			}

			go in.Engine.Run(fmt.Sprintf(":%d", in.Params.Port))
		},
	)
}
