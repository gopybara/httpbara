package main

import (
	"github.com/gopybara/httpbara"
	"go.uber.org/fx"
)

type FXHandler struct {
	fx.Out

	Handler *httpbara.Handler `group:"handlers"`
}

func AsFxHandler(h *httpbara.Handler, err ...error) (FXHandler, error) {
	if len(err) > 0 && err[0] != nil {
		return FXHandler{}, err[0]
	}

	return FXHandler{
		Handler: h,
	}, nil
}
