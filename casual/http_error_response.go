package casual

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gopybara/httpbara/common"
	"net/http"
)

type HttpErrorResponse struct {
	Status int                    `json:"status" xml:"status"`
	Error  *HttpError             `json:"error" xml:"error"`
	Meta   map[string]interface{} `json:"meta,omitempty" xml:"meta,omitempty"`
}

type HttpError struct {
	error
	frontendMessage *string
	httpCode        int

	Code    any    `json:"code,omitempty" xml:"code,omitempty"`
	Message string `json:"message" xml:"message"`

	Details []*HttpErrorField `json:"details,omitempty" xml:"details,omitempty"`
}

func (e HttpError) GetHttpStatusCode() int {
	return e.httpCode
}

func (e HttpError) GetMessage() string {
	if e.frontendMessage != nil {
		return *e.frontendMessage
	} else if e.Message != "" {
		return e.Message
	}

	return e.error.Error()
}

func (e HttpError) GetCode() any {
	return e.Code
}

var (
	ErrNotFound            = NewHTTPErrorFromMessage(http.StatusNotFound, "not found")
	ErrUnauthorized        = NewHTTPErrorFromMessage(http.StatusUnauthorized, "unauthorized")
	ErrInternalServerError = NewHTTPErrorFromMessage(http.StatusInternalServerError, "internal server error")
	ErrTooManyRequests     = NewHTTPErrorFromMessage(http.StatusTooManyRequests, "too many requests")
	ErrBadRequest          = NewHTTPErrorFromMessage(http.StatusBadRequest, "bad request")
	ErrUnprocessableEntity = NewHTTPErrorFromMessage(http.StatusUnprocessableEntity, "unprocessable entity")
)

func NewHTTPErrorFromMessage(httpCode int, message string, frontendMessage ...string) error {
	httpErr := HttpError{error: errors.New(message), httpCode: httpCode}
	if len(frontendMessage) > 0 {
		httpErr.frontendMessage = &frontendMessage[0]
	} else {
		httpErr.frontendMessage = &message
	}

	return httpErr
}

func NewHTTPErrorFromError(httpCode int, err error, frontendMessage ...string) error {
	httpErr := HttpError{error: err, httpCode: httpCode}
	if len(frontendMessage) > 0 {
		httpErr.frontendMessage = &frontendMessage[0]
	}

	return httpErr
}

func NewHttpErrorResponse(err error, opts ...HttpResponseParamsCb) (int, *HttpErrorResponse) {
	var params httpResponseParams
	params.statusCode = common.Ptr(http.StatusInternalServerError)

	for _, opt := range opts {
		opt(&params)
	}

	if params.lang == nil {
		params.lang = common.Ptr("en")
	}

	errorMessage := err.Error()

	var httpErr HttpError
	var ve validator.ValidationErrors

	if errors.As(err, &httpErr) {
		params.statusCode = common.Ptr(httpErr.GetHttpStatusCode())
		errorMessage = err.(HttpError).GetMessage()
	} else if errors.As(err, &ve) {
		for _, fe := range ve {
			params.statusCode = common.Ptr(http.StatusUnprocessableEntity)

			httpErr.Details = append(httpErr.Details, &HttpErrorField{
				Field: fe.Field(),
				Issue: getValidationErrorText(params.lang, fe),
			})
		}
	}

	httpErr.Message = errorMessage

	var metadata map[string]interface{}
	if params.meta != nil {
		metadata = params.meta
	}

	if params.statusCode == nil {
		params.statusCode = common.Ptr(http.StatusInternalServerError)
	}

	return *params.statusCode, &HttpErrorResponse{
		Status: *params.statusCode,
		Error:  &httpErr,
		Meta:   metadata,
	}
}
