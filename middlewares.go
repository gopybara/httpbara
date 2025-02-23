package httpbara

import (
	"github.com/gin-gonic/gin"
	"time"
)

type iAccessLogMiddleware struct {
	AccessLogMiddleware Middleware `middleware:"log"`
}

type accessLogMiddlewareImpl struct {
	iAccessLogMiddleware

	log Logger
}

func (alm *accessLogMiddlewareImpl) AccessLogMiddleware(ctx *gin.Context) {
	ts := time.Now()
	fields := []interface{}{
		"method", ctx.Request.Method,
		"path", ctx.Request.URL.Path,
	}
	var additionalFields []interface{}

	ctx.Set("fields", &additionalFields)

	ctx.Next()

	fields = append(fields, "status", ctx.Writer.Status())
	if len(ctx.Request.URL.Query()) > 0 {
		fields = append(fields, "query", ctx.Request.URL.Query())
	}

	fields = append(fields, "duration", time.Since(ts))

	alm.log.Info("request done", append(fields, additionalFields...)...)
}

func AddLogFieldToAccessLog(ctx *gin.Context, value ...interface{}) {
	fields, ok := ctx.Get("fields")
	if !ok {
		fields = []interface{}{}
	}

	logFields := fields.(*[]interface{})

	*logFields = append(*logFields, value...)
}

func NewAccessLogMiddleware(log Logger) *Middleware {
	alm := &accessLogMiddlewareImpl{
		log: log,
	}

	return &Middleware{
		handler:    alm.AccessLogMiddleware,
		middleware: "log",
	}
}
