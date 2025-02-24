package httpbarafx

import (
	"github.com/gopybara/httpbara"
	"go.uber.org/fx"
)

type HttpbaraOpt struct {
	fx.Out

	Opt httpbara.ParamsCb `group:"httpbaraOpts"`
}

func AsHttpbaraOpt(opt httpbara.ParamsCb) HttpbaraOpt {
	return HttpbaraOpt{
		Opt: opt,
	}
}
