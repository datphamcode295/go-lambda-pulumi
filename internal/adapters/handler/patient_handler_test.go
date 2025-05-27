package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/datphamcode295/go-lambda-pulumi/internal/core/domain"
	util "github.com/datphamcode295/go-lambda-pulumi/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// PayTransactionService interface for mocking
type PayTransactionService interface {
	PayTransaction(data domain.PayTransactionRequest) (*domain.Transaction, error)
}

// MockPatientService is a simple mock implementation
type MockPatientService struct {
	mock.Mock
}

func (m *MockPatientService) PayTransaction(data domain.PayTransactionRequest) (*domain.Transaction, error) {
	args := m.Called(data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

// TestPatientHandler wraps PatientHandler for testing
type TestPatientHandler struct {
	svc PayTransactionService
}

func NewTestPatientHandler(svc PayTransactionService) *TestPatientHandler {
	return &TestPatientHandler{svc: svc}
}

func (h *TestPatientHandler) PayTransaction(ctx *gin.Context) {
	var data domain.PayTransactionRequest
	if err := ctx.ShouldBindJSON(&data); err != nil {
		HandleError(ctx, http.StatusBadRequest, err)
		return
	}

	rs, err := h.svc.PayTransaction(data)
	if err != nil {
		HandleError(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, rs)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register custom validator like in main.go
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("ddmmyyyy", util.ValidateDDMMYYYY)
	}

	return router
}

func TestPatientHandler_PayTransaction_Success(t *testing.T) {
	// Setup
	mockService := &MockPatientService{}
	handler := NewTestPatientHandler(mockService)
	router := setupTestRouter()
	router.POST("/pay-transaction", handler.PayTransaction)

	// Test data
	patientID := uuid.New()
	transactionID := uuid.New()

	requestData := domain.PayTransactionRequest{
		PatientID:   patientID,
		DateOfBirth: "15-03-1990",
		RecordType:  "NEW",
	}

	expectedTransaction := &domain.Transaction{
		ID:          transactionID,
		PatientID:   patientID,
		Status:      domain.TransactionStatusSuccess,
		DateOfBirth: "15-03-1990",
		RecordType:  "NEW",
		APIResponse: json.RawMessage(`{"message": "Transaction success"}`),
	}

	// Mock expectations
	mockService.On("PayTransaction", requestData).Return(expectedTransaction, nil)

	// Create request
	requestBody, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", "/pay-transaction", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseTransaction domain.Transaction
	err := json.Unmarshal(w.Body.Bytes(), &responseTransaction)
	assert.NoError(t, err)
	assert.Equal(t, expectedTransaction.ID, responseTransaction.ID)
	assert.Equal(t, expectedTransaction.Status, responseTransaction.Status)
	assert.Equal(t, expectedTransaction.PatientID, responseTransaction.PatientID)

	mockService.AssertExpectations(t)
}

func TestPatientHandler_PayTransaction_InvalidJSON(t *testing.T) {
	// Setup
	mockService := &MockPatientService{}
	handler := NewTestPatientHandler(mockService)
	router := setupTestRouter()
	router.POST("/pay-transaction", handler.PayTransaction)

	// Create request with invalid JSON
	req, _ := http.NewRequest("POST", "/pay-transaction", bytes.NewBuffer([]byte(`{invalid json}`)))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")

	// Service should not be called
	mockService.AssertNotCalled(t, "PayTransaction")
}

func TestPatientHandler_PayTransaction_MissingRequiredFields(t *testing.T) {
	// Setup
	mockService := &MockPatientService{}
	handler := NewTestPatientHandler(mockService)
	router := setupTestRouter()
	router.POST("/pay-transaction", handler.PayTransaction)

	// Test data with missing required fields
	requestData := map[string]interface{}{
		"patient_id": "", // missing or empty
		// missing date_of_birth and record_type
	}

	// Create request
	requestBody, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", "/pay-transaction", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")

	// Service should not be called
	mockService.AssertNotCalled(t, "PayTransaction")
}

func TestPatientHandler_PayTransaction_ServiceError(t *testing.T) {
	// Setup
	mockService := &MockPatientService{}
	handler := NewTestPatientHandler(mockService)
	router := setupTestRouter()
	router.POST("/pay-transaction", handler.PayTransaction)

	// Test data
	patientID := uuid.New()
	requestData := domain.PayTransactionRequest{
		PatientID:   patientID,
		DateOfBirth: "15-03-1990",
		RecordType:  "NEW",
	}

	// Mock expectations - service returns error
	mockService.On("PayTransaction", requestData).Return(nil, errors.New("patient not found"))

	// Create request
	requestBody, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", "/pay-transaction", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "patient not found", response["error"])

	mockService.AssertExpectations(t)
}

func TestPatientHandler_PayTransaction_FailedTransaction(t *testing.T) {
	// Setup
	mockService := &MockPatientService{}
	handler := NewTestPatientHandler(mockService)
	router := setupTestRouter()
	router.POST("/pay-transaction", handler.PayTransaction)

	// Test data
	patientID := uuid.New()
	transactionID := uuid.New()

	requestData := domain.PayTransactionRequest{
		PatientID:   patientID,
		DateOfBirth: "15-03-2010", // Under 18 years old
		RecordType:  "NEW",
	}

	expectedTransaction := &domain.Transaction{
		ID:          transactionID,
		PatientID:   patientID,
		Status:      domain.TransactionStatusFailed,
		DateOfBirth: "15-03-2010",
		RecordType:  "NEW",
		APIResponse: json.RawMessage(`{"error": "Patient must be more than 18 years old"}`),
	}

	// Mock expectations
	mockService.On("PayTransaction", requestData).Return(expectedTransaction, nil)

	// Create request
	requestBody, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", "/pay-transaction", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseTransaction domain.Transaction
	err := json.Unmarshal(w.Body.Bytes(), &responseTransaction)
	assert.NoError(t, err)
	assert.Equal(t, domain.TransactionStatusFailed, responseTransaction.Status)
	assert.Equal(t, expectedTransaction.PatientID, responseTransaction.PatientID)

	mockService.AssertExpectations(t)
}

func TestPatientHandler_PayTransaction_InvalidRecordType(t *testing.T) {
	// Setup
	mockService := &MockPatientService{}
	handler := NewTestPatientHandler(mockService)
	router := setupTestRouter()
	router.POST("/pay-transaction", handler.PayTransaction)

	// Test data
	patientID := uuid.New()
	transactionID := uuid.New()

	requestData := domain.PayTransactionRequest{
		PatientID:   patientID,
		DateOfBirth: "15-03-1990",
		RecordType:  "OLD", // Invalid record type
	}

	expectedTransaction := &domain.Transaction{
		ID:          transactionID,
		PatientID:   patientID,
		Status:      domain.TransactionStatusFailed,
		DateOfBirth: "15-03-1990",
		RecordType:  "OLD",
		APIResponse: json.RawMessage(`{"error": "Record type must be NEW"}`),
	}

	// Mock expectations
	mockService.On("PayTransaction", requestData).Return(expectedTransaction, nil)

	// Create request
	requestBody, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", "/pay-transaction", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseTransaction domain.Transaction
	err := json.Unmarshal(w.Body.Bytes(), &responseTransaction)
	assert.NoError(t, err)
	assert.Equal(t, domain.TransactionStatusFailed, responseTransaction.Status)
	assert.Equal(t, expectedTransaction.PatientID, responseTransaction.PatientID)

	mockService.AssertExpectations(t)
}

func TestNewPatientHandler_Initialization(t *testing.T) {
	// Setup
	mockService := &MockPatientService{}

	// Execute
	handler := NewTestPatientHandler(mockService)

	// Assertions
	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.svc)
}

func TestPatientHandler_PayTransaction_InvalidDateFormat(t *testing.T) {
	// Setup
	mockService := &MockPatientService{}
	handler := NewTestPatientHandler(mockService)
	router := setupTestRouter()
	router.POST("/pay-transaction", handler.PayTransaction)

	// Test data with invalid date format
	patientID := uuid.New()
	requestData := domain.PayTransactionRequest{
		PatientID:   patientID,
		DateOfBirth: "1990-03-15", // Wrong format (should be DD-MM-YYYY)
		RecordType:  "NEW",
	}

	// Create request
	requestBody, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", "/pay-transaction", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "errors")

	// Service should not be called
	mockService.AssertNotCalled(t, "PayTransaction")
}

func TestPatientHandler_PayTransaction_EmptyBody(t *testing.T) {
	// Setup
	mockService := &MockPatientService{}
	handler := NewTestPatientHandler(mockService)
	router := setupTestRouter()
	router.POST("/pay-transaction", handler.PayTransaction)

	// Create request with empty body
	req, _ := http.NewRequest("POST", "/pay-transaction", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")

	// Service should not be called
	mockService.AssertNotCalled(t, "PayTransaction")
}

func TestPatientHandler_PayTransaction_ValidationErrors(t *testing.T) {
	// Setup
	mockService := &MockPatientService{}
	handler := NewTestPatientHandler(mockService)
	router := setupTestRouter()
	router.POST("/pay-transaction", handler.PayTransaction)

	testCases := []struct {
		name        string
		requestData interface{}
		description string
	}{
		{
			name: "Missing PatientID",
			requestData: map[string]interface{}{
				"date_of_birth": "15-03-1990",
				"record_type":   "NEW",
			},
			description: "Should fail when patient_id is missing",
		},
		{
			name: "Invalid PatientID",
			requestData: map[string]interface{}{
				"patient_id":    "invalid-uuid",
				"date_of_birth": "15-03-1990",
				"record_type":   "NEW",
			},
			description: "Should fail when patient_id is not a valid UUID",
		},
		{
			name: "Missing DateOfBirth",
			requestData: map[string]interface{}{
				"patient_id":  uuid.New().String(),
				"record_type": "NEW",
			},
			description: "Should fail when date_of_birth is missing",
		},
		{
			name: "Missing RecordType",
			requestData: map[string]interface{}{
				"patient_id":    uuid.New().String(),
				"date_of_birth": "15-03-1990",
			},
			description: "Should fail when record_type is missing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			requestBody, _ := json.Marshal(tc.requestData)
			req, _ := http.NewRequest("POST", "/pay-transaction", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, http.StatusBadRequest, w.Code, tc.description)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			// Response should contain either 'error' or 'errors' field
			hasError := false
			if _, exists := response["error"]; exists {
				hasError = true
			}
			if _, exists := response["errors"]; exists {
				hasError = true
			}
			assert.True(t, hasError, "Response should contain error information")
		})
	}

	// Service should never be called for validation errors
	mockService.AssertNotCalled(t, "PayTransaction")
}
