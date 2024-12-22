package httpbara

import (
	"github.com/gin-gonic/gin"
	"time"
)

type Params struct {
	gin             *gin.Engine
	log             Logger
	rootMiddlewares []*Middleware
	shutdownTimeout time.Duration
	taskTracker     *ActiveTaskTracker
}

type paramsCb func(*Params) error

func WithLogger(log Logger) paramsCb {
	return func(params *Params) error {
		params.log = log

		return nil
	}
}

func WithGinEngine(r *gin.Engine) paramsCb {
	return func(params *Params) error {
		params.gin = r

		return nil
	}
}

func WithRootMiddlewares(middlewares ...*Middleware) paramsCb {
	return func(params *Params) error {
		params.rootMiddlewares = middlewares

		return nil
	}
}

func WithShutdownTimeout(timeout time.Duration) paramsCb {
	return func(params *Params) error {
		params.shutdownTimeout = timeout

		return nil
	}
}

func WithTaskTracker(tracker *ActiveTaskTracker) paramsCb {
	return func(params *Params) error {
		params.taskTracker = tracker
		return nil
	}
}
