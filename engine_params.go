package httpbara

import (
	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara/casual"
	"time"
)

type params struct {
	gin             *gin.Engine
	log             Logger
	rootMiddlewares []*Middleware
	shutdownTimeout time.Duration
	taskTracker     TaskTracker

	casualResponseErrorHandler func(err error, opts ...casual.HttpResponseParamsCb) (int, interface{})
	casualResponseHandler      func(data any, opts ...casual.HttpResponseParamsCb) (int, interface{})
}

type ParamsCb func(*params) error

func WithLogger(log Logger) ParamsCb {
	return func(params *params) error {
		params.log = log

		return nil
	}
}

func WithGinEngine(r *gin.Engine) ParamsCb {
	return func(params *params) error {
		params.gin = r

		return nil
	}
}

func WithRootMiddlewares(middlewares ...*Middleware) ParamsCb {
	return func(params *params) error {
		params.rootMiddlewares = middlewares

		return nil
	}
}

func WithShutdownTimeout(timeout time.Duration) ParamsCb {
	return func(params *params) error {
		params.shutdownTimeout = timeout

		return nil
	}
}

func WithTaskTracker(tracker ...TaskTracker) ParamsCb {
	return func(params *params) error {
		if len(tracker) == 0 {
			tracker = append(tracker, NewActiveTaskTracker())
		}

		params.taskTracker = tracker[0]
		return nil
	}
}
