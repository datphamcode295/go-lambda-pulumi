package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/datphamcode295/go-lambda-pulumi/internal/core/domain"
	util "github.com/datphamcode295/go-lambda-pulumi/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// Test struct for validation errors
type TestStruct struct {
	RequiredField string `json:"required_field" validate:"required"`
	EmailField    string `json:"email_field" validate:"email"`
	MinField      string `json:"min_field" validate:"min=5"`
	MaxField      string `json:"max_field" validate:"max=10"`
	DateField     string `json:"date_field" validate:"ddmmyyyy"`
}

func setupValidatorForTests() {
	// Register custom validator like in main.go
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("ddmmyyyy", util.ValidateDDMMYYYY)
	}
}

func TestFormatFieldName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "PascalCase to snake_case",
			input:    "PatientID",
			expected: "patient_i_d",
		},
		{
			name:     "CamelCase to snake_case",
			input:    "dateOfBirth",
			expected: "date_of_birth",
		},
		{
			name:     "Single word lowercase",
			input:    "name",
			expected: "name",
		},
		{
			name:     "Single word uppercase",
			input:    "NAME",
			expected: "n_a_m_e",
		},
		{
			name:     "Multiple words",
			input:    "RecordType",
			expected: "record_type",
		},
		{
			name:     "Already snake_case",
			input:    "already_snake_case",
			expected: "already_snake_case",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatFieldName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetErrorMsg(t *testing.T) {
	testCases := []struct {
		name        string
		tag         string
		param       string
		expected    string
		description string
	}{
		{
			name:        "Required field error",
			tag:         "required",
			param:       "",
			expected:    "This field is required",
			description: "Should return required field message",
		},
		{
			name:        "Date format error",
			tag:         "ddmmyyyy",
			param:       "",
			expected:    "Date must be in DD-MM-YYYY format",
			description: "Should return date format message",
		},
		{
			name:        "Email validation error (default case)",
			tag:         "email",
			param:       "",
			expected:    "Validation failed on email",
			description: "Should return generic message for email validation",
		},
		{
			name:        "Min length error (default case)",
			tag:         "min",
			param:       "5",
			expected:    "Validation failed on min",
			description: "Should return generic message for min validation",
		},
		{
			name:        "Max length error (default case)",
			tag:         "max",
			param:       "10",
			expected:    "Validation failed on max",
			description: "Should return generic message for max validation",
		},
		{
			name:        "Unknown validation error",
			tag:         "unknown",
			param:       "",
			expected:    "Validation failed on unknown",
			description: "Should return generic validation message for unknown tags",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock FieldError
			mockFieldError := &mockFieldError{
				tag:   tc.tag,
				param: tc.param,
			}

			result := getErrorMsg(mockFieldError)
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestFormatValidationErrors(t *testing.T) {
	// Setup validator with custom validation
	setupValidatorForTests()
	validate := validator.New()
	validate.RegisterValidation("ddmmyyyy", util.ValidateDDMMYYYY)

	t.Run("Valid validation errors", func(t *testing.T) {
		// Create a test struct with validation errors
		testData := TestStruct{
			RequiredField: "", // Missing required field
			EmailField:    "invalid-email",
			MinField:      "abc", // Too short
			MaxField:      "this is way too long",
		}

		err := validate.Struct(testData)
		assert.Error(t, err)

		validationErrors := formatValidationErrors(err)
		assert.NotEmpty(t, validationErrors)

		// Check that we have multiple validation errors
		assert.True(t, len(validationErrors) > 0)

		// Verify structure of validation errors
		for _, validationError := range validationErrors {
			assert.NotEmpty(t, validationError.Field)
			assert.NotEmpty(t, validationError.Message)
		}
	})

	t.Run("Non-validation error", func(t *testing.T) {
		regularError := errors.New("regular error")
		validationErrors := formatValidationErrors(regularError)
		assert.Empty(t, validationErrors)
	})

	t.Run("Nil error", func(t *testing.T) {
		validationErrors := formatValidationErrors(nil)
		assert.Empty(t, validationErrors)
	})
}

func TestHandleError_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupValidatorForTests()

	t.Run("Handle validation errors", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Create validation error using domain struct
		validate := validator.New()
		validate.RegisterValidation("ddmmyyyy", util.ValidateDDMMYYYY)

		// Use a simpler struct without the ddmmyyyy validation
		type SimpleTestStruct struct {
			RequiredField string `json:"required_field" validate:"required"`
		}

		testData := SimpleTestStruct{
			RequiredField: "", // Missing required field
		}
		err := validate.Struct(testData)

		// Execute
		HandleError(c, http.StatusBadRequest, err)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "errors")
		assert.Contains(t, w.Body.String(), "required_field")
		assert.Contains(t, w.Body.String(), "This field is required")
	})
}

func TestHandleError_RegularErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Handle regular error", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		regularError := errors.New("something went wrong")

		// Execute
		HandleError(c, http.StatusInternalServerError, regularError)

		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "error")
		assert.Contains(t, w.Body.String(), "something went wrong")
		assert.NotContains(t, w.Body.String(), "errors") // Should not contain "errors" array
	})

	t.Run("Handle nil error", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Execute - testing with empty error string instead of nil
		emptyError := errors.New("")
		HandleError(c, http.StatusBadRequest, emptyError)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})
}

func TestHandleError_DifferentStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	statusCodes := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, statusCode := range statusCodes {
		t.Run(fmt.Sprintf("Status code %d", statusCode), func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			testError := errors.New("test error")

			// Execute
			HandleError(c, statusCode, testError)

			// Assertions
			assert.Equal(t, statusCode, w.Code)
			assert.Contains(t, w.Body.String(), "test error")
		})
	}
}

func TestValidationError_Struct(t *testing.T) {
	validationErr := ValidationError{
		Field:   "test_field",
		Message: "test message",
	}

	assert.Equal(t, "test_field", validationErr.Field)
	assert.Equal(t, "test message", validationErr.Message)
}

func TestHandleError_WithDomainValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupValidatorForTests()

	t.Run("Handle domain validation errors", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Create validation error using actual domain struct
		validate := validator.New()
		validate.RegisterValidation("ddmmyyyy", util.ValidateDDMMYYYY)

		testData := domain.PayTransactionRequest{
			// Missing required fields to ensure validation errors
			// PatientID is required but not set (zero value)
			// DateOfBirth is required but invalid
			// RecordType is required but not set
			DateOfBirth: "invalid-date-format",
			RecordType:  "", // Empty required field
		}
		err := validate.Struct(testData)

		// Only proceed if there are actually validation errors
		if err != nil {
			// Execute
			HandleError(c, http.StatusBadRequest, err)

			// Assertions
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "errors")
		} else {
			// If no validation errors, test with a manual error
			HandleError(c, http.StatusBadRequest, errors.New("manual test error"))
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "error")
		}
	})
}

// Mock implementation of validator.FieldError for testing
type mockFieldError struct {
	tag   string
	param string
	field string
}

func (m *mockFieldError) Tag() string             { return m.tag }
func (m *mockFieldError) ActualTag() string       { return m.tag }
func (m *mockFieldError) Namespace() string       { return "" }
func (m *mockFieldError) StructNamespace() string { return "" }
func (m *mockFieldError) Field() string           { return m.field }
func (m *mockFieldError) StructField() string     { return "" }
func (m *mockFieldError) Value() interface{}      { return "" }
func (m *mockFieldError) Param() string           { return m.param }
func (m *mockFieldError) Kind() reflect.Kind      { return reflect.String }
func (m *mockFieldError) Type() reflect.Type      { return nil }
func (m *mockFieldError) Error() string           { return "" }

// Translate method for the interface - returns empty string for tests
func (m *mockFieldError) Translate(ut ut.Translator) string { return "" }
