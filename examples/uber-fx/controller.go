package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara"
)

type TestHandlerDescriber struct {
	API         httpbara.Group      `group:"/api"`
	CasualRoute httpbara.Route      `route:"GET /foo" group:"api" middlewares:"logging"`
	Middleware  httpbara.Middleware `middleware:"logging"`
}

type TestHandler struct {
	TestHandlerDescriber
}

type Bar struct {
	Baz string `json:"baz"`
}

func (t *TestHandler) Middleware(ctx *gin.Context) {
	ctx.Next()
}

func (t *TestHandler) CasualRoute(ctx context.Context, req *Bar) (*Bar, error) {
	return req, nil
}

func NewTestHandler() (FXHandler, error) {
	return AsFxHandler(httpbara.AsHandler(&TestHandler{}))
}
