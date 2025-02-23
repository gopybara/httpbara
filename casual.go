package httpbara

import (
	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara/casual"
	"reflect"
)

type casualRoute struct {
	middlewares []string
	group       string
	method      string
	path        string
	handler     *casualHandler
}

type casualHandler struct {
	rv *reflect.Value
	rm *reflect.Method
}

func isCasualHandler(t reflect.Type) bool {
	if t.NumIn() != 3 ||
		t.NumOut() < 1 {
		return false
	}

	if t.In(1).String() != reflect.TypeOf((*gin.Context)(nil)).String() && t.In(1).String() != "context.Context" {
		return false
	}

	switch t.NumOut() {
	case 1:
		return t.Out(0).String() == "error"
	case 2:
		return t.Out(1).String() == "error"
	default:
		return false
	}
}

// Basic casual responses
func defaultCasualErrorResponder(err error, opts ...casual.HttpResponseParamsCb) (int, interface{}) {
	return casual.NewHttpErrorResponse(err, opts...)
}

func defaultCasualResponder[T any](value T, opts ...casual.HttpResponseParamsCb) (int, any) {
	return casual.NewHTTPResponse[T](&value, opts...)
}
