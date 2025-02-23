package main

import (
	"github.com/gin-gonic/gin"
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

func (t *TestHandler) CasualRoute(ctx *gin.Context, req *Bar) (*Bar, error) {
	httpbara.AddLogFieldToAccessLog(ctx, "foo", "bar")

	return req, nil
}

func NewTestHandler() (FXHandler, error) {
	return AsFxHandler(httpbara.AsHandler(&TestHandler{}))
}
