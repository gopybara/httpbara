package casual

import (
	"github.com/gopybara/httpbara/common"
	"net/http"
	"reflect"
)

type HttpResponse[T any] struct {
	Status int                    `json:"status" xml:"status"`
	Data   *T                     `json:"data,omitempty" xml:"data,omitempty"`
	Meta   map[string]interface{} `json:"meta,omitempty" xml:"meta,omitempty"`
}

func NewHTTPResponse[T any](data *T, opts ...HttpResponseParamsCb) (int, *HttpResponse[T]) {
	var params httpResponseParams
	for _, opt := range opts {
		opt(&params)
	}

	elem := reflect.ValueOf(data)
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}

	var metadata map[string]interface{}
	if params.meta != nil {
		metadata = params.meta
	}

	if data != nil && (elem.Kind() == reflect.Slice || elem.Kind() == reflect.Array) {
		if metadata == nil {
			metadata = make(map[string]interface{})
		}

		if _, ok := metadata["total"]; !ok {
			metadata["total"] = elem.Len()
		}
	}

	if params.statusCode == nil {
		params.statusCode = common.Ptr(http.StatusOK)
	}

	return *params.statusCode, &HttpResponse[T]{
		Status: *params.statusCode,
		Data:   data,
		Meta:   metadata,
	}
}
