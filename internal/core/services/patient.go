package services

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/datphamcode295/go-lambda-pulumi/internal/config"
	"github.com/datphamcode295/go-lambda-pulumi/internal/core/domain"
	"github.com/datphamcode295/go-lambda-pulumi/internal/core/ports"
	"github.com/google/uuid"
)

type PatientService struct {
	cfg             *config.Config
	patientRepo     ports.PatientRepository
	transactionRepo ports.TransactionRepository
}

func NewPatientService(cfg *config.Config, patientRepo ports.PatientRepository, transactionRepo ports.TransactionRepository) *PatientService {
	return &PatientService{
		cfg:             cfg,
		patientRepo:     patientRepo,
		transactionRepo: transactionRepo,
	}
}

func (p *PatientService) PayTransaction(data domain.PayTransactionRequest) (*domain.Transaction, error) {
	patient, err := p.patientRepo.GetPatient(data.PatientID.String())
	if err != nil {
		return nil, err
	}

	transaction := domain.Transaction{
		ID:          uuid.New(),
		PatientID:   data.PatientID,
		DateOfBirth: data.DateOfBirth,
		RecordType:  data.RecordType,
	}

	// validate data
	// patient more than 18 years old
	patientDateOfBirth, err := time.Parse("02-01-2006", data.DateOfBirth)
	if err != nil {
		return nil, errors.New("date of birth format must be DD-MM-YYYY")
	}

	patientAge := math.Floor(time.Since(patientDateOfBirth).Hours() / 24 / 365)
	if patientAge < 18 {
		transaction.Status = domain.TransactionStatusFailed
		transaction.APIResponse = json.RawMessage(`{"error": "Patient must be more than 18 years old"}`)
		rs, err := p.transactionRepo.CreateTransaction(transaction)
		if err != nil {
			return nil, err
		}
		return rs, nil
	}

	// only accept record with type NEW
	if data.RecordType != "NEW" {
		transaction.Status = domain.TransactionStatusFailed
		transaction.APIResponse = json.RawMessage(`{"error": "Record type must be NEW"}`)
		rs, err := p.transactionRepo.CreateTransaction(transaction)
		if err != nil {
			return nil, err
		}

		return rs, nil
	}

	// remap
	type SubmitPatientRequest struct {
		Patient    *domain.Patient `json:"patient"`
		Age        int             `json:"age"`
		RecordType string          `json:"record_type"`
	}

	submitPatientRequest := SubmitPatientRequest{
		Patient:    patient,
		Age:        int(patientAge),
		RecordType: data.RecordType,
	}

	// call external api
	fmt.Printf("Calling external api with request: %+v\n and api key: %s\n", submitPatientRequest, p.cfg.APIKey)
	isSuccess, err := rand.Int(rand.Reader, big.NewInt(2))
	if err != nil {
		return nil, err
	}

	// if isSuccess is 0 means transaction failed
	if isSuccess.Int64() == 0 {
		dummyFailedBody := map[string]string{
			"error": "Transaction failed",
		}

		jsonBody, err := json.Marshal(dummyFailedBody)
		if err != nil {
			return nil, err
		}

		transaction.Status = domain.TransactionStatusFailed
		transaction.APIResponse = jsonBody
		rs, err := p.transactionRepo.CreateTransaction(transaction)
		if err != nil {
			return nil, err
		}
		return rs, nil
	} else {
		// create transaction with status success
		dummySuccessBody := map[string]string{
			"message": "Transaction success",
		}
		jsonBody, err := json.Marshal(dummySuccessBody)
		if err != nil {
			return nil, err
		}

		transaction.Status = domain.TransactionStatusSuccess
		transaction.APIResponse = jsonBody
		rs, err := p.transactionRepo.CreateTransaction(transaction)
		if err != nil {
			return nil, err
		}
		return rs, nil
	}
}
