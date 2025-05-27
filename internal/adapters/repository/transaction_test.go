package repository_test

import (
	"testing"
	"time"

	"github.com/datphamcode295/go-lambda-pulumi/internal/adapters/repository"
	"github.com/datphamcode295/go-lambda-pulumi/internal/core/domain"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
)

// setupTestDB initializes a new in-memory SQLite database for testing.
func setupTestDBForTransaction() (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	// Auto-migrate schemas for Patient and Transaction
	db.AutoMigrate(&domain.Patient{})
	db.AutoMigrate(&domain.Transaction{})
	return db, nil
}

func TestCreateTransaction(t *testing.T) {
	db, err := setupTestDBForTransaction()
	assert.NoError(t, err)

	repo := repository.NewDB(db)

	// Prepare a patient for the transaction
	patientID := uuid.New()
	patient := &domain.Patient{ID: patientID, Name: "Test Patient For Transaction"}
	db.Create(patient)

	// Case 1: Successfully create a transaction
	transactionID := uuid.New()
	transactionToCreate := domain.Transaction{
		ID:        transactionID,
		PatientID: patientID,
		Status:    domain.TransactionStatusSuccess,
		CreatedAt: time.Now(),
	}

	createdTransaction, err := repo.CreateTransaction(transactionToCreate)
	assert.NoError(t, err)
	assert.NotNil(t, createdTransaction)
	assert.Equal(t, transactionToCreate.ID, createdTransaction.ID)
	assert.Equal(t, transactionToCreate.PatientID, createdTransaction.PatientID)
	assert.Equal(t, transactionToCreate.Status, createdTransaction.Status)

	// Verify that the transaction is actually in the database
	var fetchedTransaction domain.Transaction
	db.First(&fetchedTransaction, "id = ?", transactionID)
	assert.Equal(t, transactionToCreate.ID, fetchedTransaction.ID)
}
