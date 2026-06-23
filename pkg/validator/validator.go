package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

var validate = validator.New()

func init() {
	validate.RegisterValidation("currency", validateCurrency)
	validate.RegisterValidation("decimal_gt", validateDecimalGT)
}

func Struct(s interface{}) error {
	return validate.Struct(s)
}

func Var(field interface{}, tag string) error {
	return validate.Var(field, tag)
}

func validateCurrency(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	valid := map[string]bool{
		"USD": true, "EUR": true, "GBP": true, "CAD": true,
		"AUD": true, "JPY": true, "HKD": true,
	}
	return valid[v]
}

func validateDecimalGT(fl validator.FieldLevel) bool {
	val, _ := decimal.NewFromString(fl.Field().String())
	param, _ := decimal.NewFromString(fl.Param())
	return val.GreaterThan(param)
}
