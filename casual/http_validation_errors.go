package casual

import (
	"github.com/go-playground/validator/v10"
	"strings"
)

type HttpErrorField struct {
	Field string `json:"field" xml:"field"`
	Issue string `json:"issue" xml:"issue"`
}

var validationErrors = map[string]func(lang *string, fe validator.FieldError) string{
	"required": func(lang *string, fe validator.FieldError) string {
		return "Field is required"
	},
	"lte": func(lang *string, fe validator.FieldError) string {
		return "Should be less than " + fe.Param()
	},
	"gte": func(lang *string, fe validator.FieldError) string {
		return "Should be greater than " + fe.Param()
	},
	"oneof": func(lang *string, fe validator.FieldError) string {
		return "Should be one of [" + strings.Join(strings.Split(fe.Param(), " "), ",") + "]"
	},
	"notempty": func(lang *string, fe validator.FieldError) string {
		return "Param should not be empty"
	},
	"email": func(lang *string, fe validator.FieldError) string {
		return "Param should be valid email " + fe.Param()
	},
	"url": func(lang *string, fe validator.FieldError) string {
		return "Param should be valid url"
	},
	"min": func(lang *string, fe validator.FieldError) string {
		return "Param should be greater than " + fe.Param()
	},
	"max": func(lang *string, fe validator.FieldError) string {
		return "Param should be less than " + fe.Param()
	},
}

func getValidationErrorText(lang *string, fe validator.FieldError) string {
	if msg, ok := validationErrors[fe.Tag()]; ok {
		return msg(lang, fe)
	}

	return "Unknown error"
}

type ValidationErrorMessageFunc func(lang *string, fe validator.FieldError) string

func DefaultValidationErrorMessageFunc(message string) ValidationErrorMessageFunc {
	return func(lang *string, fe validator.FieldError) string {
		return message
	}
}

func AddValidationErrorMessage(tag string, messageFunc ValidationErrorMessageFunc) {
	validationErrors[tag] = messageFunc
}

func AddValidationErrorMessages(errors map[string]ValidationErrorMessageFunc) {
	for key, value := range errors {
		validationErrors[key] = value
	}
}
