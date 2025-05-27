package handler

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func formatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			var element ValidationError
			element.Field = formatFieldName(e.Field())
			element.Message = getErrorMsg(e)
			errors = append(errors, element)
		}
	}

	return errors
}

func formatFieldName(field string) string {
	// Convert PascalCase to snake_case
	var result []rune
	for i, r := range field {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

func getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "ddmmyyyy":
		return "Date must be in DD-MM-YYYY format"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Minimum length is %s", fe.Param())
	case "max":
		return fmt.Sprintf("Maximum length is %s", fe.Param())
	default:
		return fmt.Sprintf("Validation failed on %s", fe.Tag())
	}
}

func HandleError(ctx *gin.Context, statusCode int, err error) {
	// check if it is a validation error
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		ctx.JSON(statusCode, gin.H{
			"errors": formatValidationErrors(validationErrors),
		})
		return
	}

	ctx.JSON(statusCode, gin.H{
		"error": err.Error(),
	})
}
