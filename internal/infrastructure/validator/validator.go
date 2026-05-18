package validator

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	domainerror "github.com/rizkicandra/dandanna-api/internal/domain/error"
)

var v *validator.Validate

func init() {
	v = validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})
}

// Struct validates s and returns a *domainerror.Validation for the first failing field,
// or nil if all fields pass.
func Struct(s interface{}) error {
	if err := v.Struct(s); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			fe := ve[0]
			return &domainerror.Validation{
				Field:   fe.Field(),
				Code:    codeFor(fe.Field(), fe.Tag()),
				Message: messageFor(fe.Field(), fe.Tag(), fe.Param()),
			}
		}
		return err
	}
	return nil
}

func codeFor(field, tag string) string {
	switch tag {
	case "required":
		return "REQUIRED"
	case "email":
		return "INVALID_EMAIL"
	case "min":
		return "TOO_SHORT"
	case "max":
		return "TOO_LONG"
	case "e164":
		return "INVALID_PHONE"
	case "oneof":
		switch field {
		case "primary_service":
			return "INVALID_SERVICE"
		}
		return "INVALID_VALUE"
	default:
		return "INVALID"
	}
}

func messageFor(field, tag, param string) string {
	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return field + " is not valid"
	case "e164":
		return field + " must be in E.164 format (e.g. +6281234567890)"
	case "min":
		return field + " must be at least " + param + " characters"
	case "max":
		return field + " must be at most " + param + " characters"
	case "oneof":
		return field + " must be one of: " + strings.ReplaceAll(param, " ", ", ")
	default:
		return field + " is invalid"
	}
}
