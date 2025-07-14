package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the validator instance
type Validator struct {
	validator *validator.Validate
}

// New creates a new validator instance
func New() *Validator {
	v := validator.New()

	// Register custom tag name function
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validator: v}
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}
