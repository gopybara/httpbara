package main

import (
	"context"
	"github.com/gopybara/httpbara"
)

type TestHandlerDescriber struct {
	CasualRoute httpbara.Route `route:"GET /foo"`
}

type TestHandler struct {
	TestHandlerDescriber
}

type Bar struct {
	Baz string
}

func (t *TestHandler) CasualRoute(ctx context.Context, req *Bar) (*Bar, error) {
	return req, nil
}

func NewTestHandler() (*httpbara.Handler, error) {
	return httpbara.AsHandler(&TestHandler{})
}
