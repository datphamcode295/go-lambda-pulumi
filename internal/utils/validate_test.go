package util

import (
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// MockFieldLevel implements validator.FieldLevel for testing
type MockFieldLevel struct {
	value reflect.Value
}

func (m *MockFieldLevel) Top() reflect.Value      { return reflect.Value{} }
func (m *MockFieldLevel) Parent() reflect.Value   { return reflect.Value{} }
func (m *MockFieldLevel) Field() reflect.Value    { return m.value }
func (m *MockFieldLevel) FieldName() string       { return "test_field" }
func (m *MockFieldLevel) StructFieldName() string { return "TestField" }
func (m *MockFieldLevel) Param() string           { return "" }
func (m *MockFieldLevel) GetTag() string          { return "ddmmyyyy" }
func (m *MockFieldLevel) ExtractType(field reflect.Value) (reflect.Value, reflect.Kind, bool) {
	return field, field.Kind(), true
}
func (m *MockFieldLevel) GetStructFieldOK() (reflect.Value, reflect.Kind, bool) {
	return reflect.Value{}, reflect.Invalid, false
}
func (m *MockFieldLevel) GetStructFieldOKAdvanced(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool) {
	return reflect.Value{}, reflect.Invalid, false
}
func (m *MockFieldLevel) GetStructFieldOK2() (reflect.Value, reflect.Kind, bool, bool) {
	return reflect.Value{}, reflect.Invalid, false, false
}
func (m *MockFieldLevel) GetStructFieldOKAdvanced2(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool, bool) {
	return reflect.Value{}, reflect.Invalid, false, false
}

// Helper function to create a MockFieldLevel with a string value
func createMockFieldLevel(value string) validator.FieldLevel {
	return &MockFieldLevel{
		value: reflect.ValueOf(value),
	}
}

func TestValidateDDMMYYYY_ValidDates(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid date - typical",
			input:    "15-03-1990",
			expected: true,
		},
		{
			name:     "Valid date - leap year",
			input:    "29-02-2020",
			expected: true,
		},
		{
			name:     "Valid date - first day of year",
			input:    "01-01-2000",
			expected: true,
		},
		{
			name:     "Valid date - last day of year",
			input:    "31-12-2023",
			expected: true,
		},
		{
			name:     "Valid date - February non-leap year",
			input:    "28-02-2021",
			expected: true,
		},
		{
			name:     "Valid date - 30-day month",
			input:    "30-04-2022",
			expected: true,
		},
		{
			name:     "Valid date - 31-day month",
			input:    "31-01-2022",
			expected: true,
		},
		{
			name:     "Valid date - future date",
			input:    "25-12-2030",
			expected: true,
		},
		{
			name:     "Valid date - old date",
			input:    "01-01-1900",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fieldLevel := createMockFieldLevel(tc.input)
			result := ValidateDDMMYYYY(fieldLevel)
			assert.Equal(t, tc.expected, result, "Expected %v for input: %s", tc.expected, tc.input)
		})
	}
}

func TestValidateDDMMYYYY_InvalidFormats(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Wrong format - YYYY-MM-DD",
			input:    "1990-03-15",
			expected: false,
		},
		{
			name:     "Wrong format - MM-DD-YYYY",
			input:    "03-15-1990",
			expected: false,
		},
		{
			name:     "Wrong format - DD/MM/YYYY",
			input:    "15/03/1990",
			expected: false,
		},
		{
			name:     "Wrong format - DD.MM.YYYY",
			input:    "15.03.1990",
			expected: false,
		},
		{
			name:     "Wrong format - no separators",
			input:    "15031990",
			expected: false,
		},
		{
			name:     "Wrong format - single digit day",
			input:    "5-03-1990",
			expected: false,
		},
		{
			name:     "Wrong format - single digit month",
			input:    "15-3-1990",
			expected: false,
		},
		{
			name:     "Wrong format - two-digit year",
			input:    "15-03-90",
			expected: false,
		},
		{
			name:     "Wrong format - extra characters",
			input:    "15-03-1990a",
			expected: false,
		},
		{
			name:     "Wrong format - leading zeros missing",
			input:    "1-1-1990",
			expected: false,
		},
		{
			name:     "Wrong format - spaces",
			input:    "15 - 03 - 1990",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Only dashes",
			input:    "--",
			expected: false,
		},
		{
			name:     "Too many dashes",
			input:    "15-03-19-90",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fieldLevel := createMockFieldLevel(tc.input)
			result := ValidateDDMMYYYY(fieldLevel)
			assert.Equal(t, tc.expected, result, "Expected %v for input: %s", tc.expected, tc.input)
		})
	}
}

func TestValidateDDMMYYYY_InvalidDates(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Invalid day - 32nd day",
			input:    "32-01-2022",
			expected: false,
		},
		{
			name:     "Invalid day - 0th day",
			input:    "00-01-2022",
			expected: false,
		},
		{
			name:     "Invalid month - 13th month",
			input:    "15-13-2022",
			expected: false,
		},
		{
			name:     "Invalid month - 0th month",
			input:    "15-00-2022",
			expected: false,
		},
		{
			name:     "Invalid February date - 29th in non-leap year",
			input:    "29-02-2021",
			expected: false,
		},
		{
			name:     "Invalid February date - 30th",
			input:    "30-02-2020",
			expected: false,
		},
		{
			name:     "Invalid April date - 31st",
			input:    "31-04-2022",
			expected: false,
		},
		{
			name:     "Invalid June date - 31st",
			input:    "31-06-2022",
			expected: false,
		},
		{
			name:     "Invalid September date - 31st",
			input:    "31-09-2022",
			expected: false,
		},
		{
			name:     "Invalid November date - 31st",
			input:    "31-11-2022",
			expected: false,
		},
		{
			name:     "Non-numeric day",
			input:    "aa-03-1990",
			expected: false,
		},
		{
			name:     "Non-numeric month",
			input:    "15-bb-1990",
			expected: false,
		},
		{
			name:     "Non-numeric year",
			input:    "15-03-cccc",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fieldLevel := createMockFieldLevel(tc.input)
			result := ValidateDDMMYYYY(fieldLevel)
			assert.Equal(t, tc.expected, result, "Expected %v for input: %s", tc.expected, tc.input)
		})
	}
}

func TestValidateDDMMYYYY_LeapYearEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Leap year - 2020 (divisible by 4)",
			input:    "29-02-2020",
			expected: true,
		},
		{
			name:     "Non-leap year - 2021 (not divisible by 4)",
			input:    "29-02-2021",
			expected: false,
		},
		{
			name:     "Leap year - 2000 (divisible by 400)",
			input:    "29-02-2000",
			expected: true,
		},
		{
			name:     "Non-leap year - 1900 (divisible by 100 but not 400)",
			input:    "29-02-1900",
			expected: false,
		},
		{
			name:     "Leap year - 1600 (divisible by 400)",
			input:    "29-02-1600",
			expected: true,
		},
		{
			name:     "Non-leap year - 1700 (divisible by 100 but not 400)",
			input:    "29-02-1700",
			expected: false,
		},
		{
			name:     "Valid February 28 in non-leap year",
			input:    "28-02-2021",
			expected: true,
		},
		{
			name:     "Valid February 28 in leap year",
			input:    "28-02-2020",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fieldLevel := createMockFieldLevel(tc.input)
			result := ValidateDDMMYYYY(fieldLevel)
			assert.Equal(t, tc.expected, result, "Expected %v for input: %s", tc.expected, tc.input)
		})
	}
}

func TestValidateDDMMYYYY_BoundaryValues(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Minimum valid date",
			input:    "01-01-0001",
			expected: true,
		},
		{
			name:     "Maximum days in January",
			input:    "31-01-2022",
			expected: true,
		},
		{
			name:     "Maximum days in March",
			input:    "31-03-2022",
			expected: true,
		},
		{
			name:     "Maximum days in May",
			input:    "31-05-2022",
			expected: true,
		},
		{
			name:     "Maximum days in July",
			input:    "31-07-2022",
			expected: true,
		},
		{
			name:     "Maximum days in August",
			input:    "31-08-2022",
			expected: true,
		},
		{
			name:     "Maximum days in October",
			input:    "31-10-2022",
			expected: true,
		},
		{
			name:     "Maximum days in December",
			input:    "31-12-2022",
			expected: true,
		},
		{
			name:     "Maximum days in April (30)",
			input:    "30-04-2022",
			expected: true,
		},
		{
			name:     "Maximum days in June (30)",
			input:    "30-06-2022",
			expected: true,
		},
		{
			name:     "Maximum days in September (30)",
			input:    "30-09-2022",
			expected: true,
		},
		{
			name:     "Maximum days in November (30)",
			input:    "30-11-2022",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fieldLevel := createMockFieldLevel(tc.input)
			result := ValidateDDMMYYYY(fieldLevel)
			assert.Equal(t, tc.expected, result, "Expected %v for input: %s", tc.expected, tc.input)
		})
	}
}

func TestValidateDDMMYYYY_SpecialCharacters(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Unicode characters",
			input:    "１５-０３-１９９０",
			expected: false,
		},
		{
			name:     "Mixed valid and invalid separators",
			input:    "15/03-1990",
			expected: false,
		},
		{
			name:     "Tabs as separators",
			input:    "15	03	1990",
			expected: false,
		},
		{
			name:     "Multiple dashes",
			input:    "15--03--1990",
			expected: false,
		},
		{
			name:     "Leading spaces",
			input:    " 15-03-1990",
			expected: false,
		},
		{
			name:     "Trailing spaces",
			input:    "15-03-1990 ",
			expected: false,
		},
		{
			name:     "Newline characters",
			input:    "15-03-1990\n",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fieldLevel := createMockFieldLevel(tc.input)
			result := ValidateDDMMYYYY(fieldLevel)
			assert.Equal(t, tc.expected, result, "Expected %v for input: %s", tc.expected, tc.input)
		})
	}
}

func TestValidateDDMMYYYY_RegexPatternValidation(t *testing.T) {
	// Test cases specifically for regex pattern validation
	testCases := []struct {
		name        string
		input       string
		expected    bool
		description string
	}{
		{
			name:        "Correct format passes regex",
			input:       "15-03-1990",
			expected:    true,
			description: "Should pass regex validation for DD-MM-YYYY",
		},
		{
			name:        "Three digit day fails regex",
			input:       "015-03-1990",
			expected:    false,
			description: "Should fail regex validation for DDD-MM-YYYY",
		},
		{
			name:        "Three digit month fails regex",
			input:       "15-003-1990",
			expected:    false,
			description: "Should fail regex validation for DD-MMM-YYYY",
		},
		{
			name:        "Five digit year fails regex",
			input:       "15-03-19901",
			expected:    false,
			description: "Should fail regex validation for DD-MM-YYYYY",
		},
		{
			name:        "Three digit year fails regex",
			input:       "15-03-199",
			expected:    false,
			description: "Should fail regex validation for DD-MM-YYY",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fieldLevel := createMockFieldLevel(tc.input)
			result := ValidateDDMMYYYY(fieldLevel)
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

// Integration test using the actual validator package
func TestValidateDDMMYYYY_WithActualValidator(t *testing.T) {
	// Create a validator instance
	validate := validator.New()

	// Register our custom validation
	validate.RegisterValidation("ddmmyyyy", ValidateDDMMYYYY)

	// Test struct for validation
	type TestStruct struct {
		DateField string `validate:"ddmmyyyy"`
	}

	testCases := []struct {
		name      string
		dateValue string
		shouldErr bool
	}{
		{
			name:      "Valid date should pass",
			dateValue: "15-03-1990",
			shouldErr: false,
		},
		{
			name:      "Invalid format should fail",
			dateValue: "1990-03-15",
			shouldErr: true,
		},
		{
			name:      "Invalid date should fail",
			dateValue: "32-01-2022",
			shouldErr: true,
		},
		{
			name:      "Empty string should fail",
			dateValue: "",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testStruct := TestStruct{
				DateField: tc.dateValue,
			}

			err := validate.Struct(testStruct)

			if tc.shouldErr {
				assert.Error(t, err, "Expected validation to fail for: %s", tc.dateValue)
			} else {
				assert.NoError(t, err, "Expected validation to pass for: %s", tc.dateValue)
			}
		})
	}
}

// Benchmark test to check performance
func BenchmarkValidateDDMMYYYY(b *testing.B) {
	fieldLevel := createMockFieldLevel("15-03-1990")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateDDMMYYYY(fieldLevel)
	}
}

// Test for potential panic scenarios
func TestValidateDDMMYYYY_NoPanic(t *testing.T) {
	testCases := []string{
		"",
		"invalid",
		"15-03-1990",
		"99-99-9999",
		"ab-cd-efgh",
		"15-03-1990-extra",
		"15/03/1990",
	}

	for _, tc := range testCases {
		t.Run("No panic for: "+tc, func(t *testing.T) {
			fieldLevel := createMockFieldLevel(tc)

			// This should not panic
			assert.NotPanics(t, func() {
				ValidateDDMMYYYY(fieldLevel)
			}, "ValidateDDMMYYYY should not panic for input: %s", tc)
		})
	}
}
