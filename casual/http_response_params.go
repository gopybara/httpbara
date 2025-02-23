package casual

type httpResponseParams struct {
	statusCode *int
	meta       map[string]interface{}
	lang       *string
}

type HttpResponseParamsCb func(params *httpResponseParams)

func WithHttpStatusCode(code int) HttpResponseParamsCb {
	return func(params *httpResponseParams) {
		params.statusCode = &code
	}
}

func WithLang(lang string) HttpResponseParamsCb {
	return func(params *httpResponseParams) {
		params.lang = &lang
	}
}

func WithMeta(meta map[string]interface{}) HttpResponseParamsCb {
	return func(params *httpResponseParams) {
		params.meta = meta
	}
}
