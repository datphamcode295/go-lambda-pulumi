package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         string `json:"id" db:"id"`
	Email      string `json:"email" db:"email"`
	Password   string `json:"password" db:"password"`
	Membership bool   `json:"membership" db:"membership"`
}

type Patient struct {
	ID      uuid.UUID `json:"id" db:"id"`
	Name    string    `json:"name" db:"name"`
	Email   string    `json:"email" db:"email"`
	Phone   string    `json:"phone" db:"phone"`
	Address string    `json:"address" db:"address"`
	City    string    `json:"city" db:"city"`
	State   string    `json:"state" db:"state"`
	Zip     string    `json:"zip" db:"zip"`
}

type TransactionStatus string

const (
	TransactionStatusSuccess TransactionStatus = "success"
	TransactionStatusFailed  TransactionStatus = "failed"
)

type Transaction struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	PatientID   uuid.UUID         `json:"patient_id" db:"patient_id"`
	Status      TransactionStatus `json:"status" db:"status"`
	APIResponse json.RawMessage   `json:"api_response" db:"api_response"`
	RecordType  string            `json:"record_type" db:"record_type"`
	DateOfBirth string            `json:"date_of_birth" db:"date_of_birth"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
}

type PayTransactionRequest struct {
	PatientID   uuid.UUID `json:"patient_id" binding:"required"`
	DateOfBirth string    `json:"date_of_birth" binding:"required,ddmmyyyy"` // with format DD-MM-YYYY
	RecordType  string    `json:"record_type" binding:"required"`
}
