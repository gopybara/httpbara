package httpbara

import (
	"github.com/gin-gonic/gin"
	"time"
)

type IAccessLogMiddleware struct {
	AccessLogMiddleware Middleware `middleware:"log"`
}

type AccessLogMiddlewareImpl struct {
	IAccessLogMiddleware

	log Logger
}

func (alm *AccessLogMiddlewareImpl) AccessLogMiddleware(ctx *gin.Context) {
	ts := time.Now()
	ctx.Set("fields", make([]interface{}, 0))

	ctx.Next()

	alm.log.Info("request done", "method", ctx.Request.Method, "path", ctx.Request.URL.Path, "query", ctx.Request.URL.RawQuery, "took", time.Since(ts))
}

func NewAccessLogMiddleware(log Logger) *Middleware {
	alm := &AccessLogMiddlewareImpl{
		log: log,
	}

	return &Middleware{
		handler:    alm.AccessLogMiddleware,
		middleware: "log",
	}
}
