package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/go_backend_misc/util"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		// Check if the currency is supported
		return util.IsSupportedCurrency(currency)
	}
	return false
}
