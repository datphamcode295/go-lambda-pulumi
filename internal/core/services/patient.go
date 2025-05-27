package services

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
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

	// remap
	type RemapRequest struct {
		Patient     *domain.Patient `json:"patient"`
		DateOfBirth string          `json:"date_of_birth"`
		RecordType  string          `json:"record_type"`
	}

	remapRequest := RemapRequest{
		Patient:     patient,
		DateOfBirth: data.DateOfBirth,
		RecordType:  data.RecordType,
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

	patientAge := time.Since(patientDateOfBirth).Hours() / 24 / 365
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

	// call external api
	// TODO: get api key from config
	fmt.Printf("Calling external api with request: %+v\n and api key: %s\n", remapRequest, p.cfg.APIKey)
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

		// create transaction with status failed
		// transaction := domain.Transaction{
		// 	ID:           uuid.New(),
		// 	PatientID:    data.PatientID,
		// 	Status:       domain.TransactionStatusFailed,
		// 	FailedReason: &failReason,
		// }
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
