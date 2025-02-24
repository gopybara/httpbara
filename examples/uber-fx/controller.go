package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara"
)

type TestHandlerDescriber struct {
	Xyu         httpbara.Group      `group:"/xyu"`
	CasualRoute httpbara.Route      `route:"GET /foo" group:"xyu" middlewares:"loh"`
	Middleware  httpbara.Middleware `middleware:"loh"`
}

type TestHandler struct {
	TestHandlerDescriber
}

type Bar struct {
	Baz string `json:"baz"`
}

func (t *TestHandler) Middleware(ctx *gin.Context) {
	panic("Мамку ебал")
}

func (t *TestHandler) CasualRoute(ctx context.Context, req *Bar) (*Bar, error) {
	return req, nil
}

func NewTestHandler() (FXHandler, error) {
	return AsFxHandler(httpbara.AsHandler(&TestHandler{}))
}
