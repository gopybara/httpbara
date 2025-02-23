package httpbara

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara/casual"
)

var (
	ErrTaskTrackerNotSet = errors.New("task tracker is not set")
	ErrLoggerNotSet      = errors.New("logger is not set")

	ErrShutdown = casual.NewHTTPErrorFromMessage(503, "server is shutting down")
)

type taskTrackerMiddlewareDescriber struct {
	Middleware Middleware `middleware:"taskTracker"`
}

type taskTrackerMiddleware struct {
	taskTrackerMiddlewareDescriber

	log Logger
	tt  TaskTracker
}

func NewTaskTrackerMiddleware(log Logger, tt TaskTracker) (*Handler, error) {
	if tt == nil {
		return nil, ErrTaskTrackerNotSet
	}

	if log == nil {
		return nil, ErrLoggerNotSet
	}

	ttmw := taskTrackerMiddleware{
		tt:  tt,
		log: log,
	}

	return AsHandler(&ttmw)
}

func (ttmw *taskTrackerMiddleware) Middleware(ctx *gin.Context) {
	err := ttmw.tt.StartTask()
	if err != nil {
		ttmw.log.Error("cannot handle request: server is shutting down", "error", err)
		ctx.JSON(casual.NewHttpErrorResponse(ErrShutdown))
		return
	}

	defer ttmw.tt.FinishTask()

	ctx.Next()
}
