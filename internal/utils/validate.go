package util

import (
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

func ValidateDDMMYYYY(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()

	// Check format with regex
	matched, _ := regexp.MatchString(`^\d{2}-\d{2}-\d{4}$`, dateStr)
	if !matched {
		return false
	}

	// Parse and validate actual date
	_, err := time.Parse("02-01-2006", dateStr)
	return err == nil
}
