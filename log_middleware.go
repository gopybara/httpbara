package httpbara

import (
	"github.com/gin-gonic/gin"
	"time"
)

type accessLogMiddlewareDescriber struct {
	AccessLogMiddleware Middleware `middleware:"log"`
}

type accessLogMiddleware struct {
	accessLogMiddlewareDescriber

	log Logger
}

func (alm *accessLogMiddleware) AccessLogMiddleware(ctx *gin.Context) {
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

func NewAccessLogMiddleware(log Logger) (*Handler, error) {
	alm := accessLogMiddleware{
		log: log,
	}

	return AsHandler(&alm)
}
