package ports

import (
	"github.com/datphamcode295/go-lambda-pulumi/internal/core/domain"
)

type PatientService interface {
	PayTransaction(data domain.PayTransactionRequest) (*domain.Transaction, error)
}

type PatientRepository interface {
	GetPatient(id string) (*domain.Patient, error)
}

type TransactionRepository interface {
	CreateTransaction(transaction domain.Transaction) (*domain.Transaction, error)
}
