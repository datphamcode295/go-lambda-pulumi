package repository_test

import (
	"testing"

	"errors"

	"github.com/datphamcode295/go-lambda-pulumi/internal/adapters/repository"
	"github.com/datphamcode295/go-lambda-pulumi/internal/core/domain"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
)

func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	// Auto-migrate schema
	db.AutoMigrate(&domain.Patient{})
	return db, nil
}

func TestGetPatient(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	repo := repository.NewDB(db)

	// Case 1: Patient exists
	patientID := uuid.New()
	existingPatient := &domain.Patient{ID: patientID, Name: "Test Patient"}
	db.Create(existingPatient)

	foundPatient, err := repo.GetPatient(patientID.String())
	assert.NoError(t, err)
	assert.NotNil(t, foundPatient)
	assert.Equal(t, existingPatient.ID, foundPatient.ID)
	assert.Equal(t, existingPatient.Name, foundPatient.Name)

	// Case 2: Patient does not exist
	notFoundPatient, err := repo.GetPatient(uuid.New().String())
	assert.Error(t, err)
	assert.Nil(t, notFoundPatient)
	assert.Equal(t, errors.New("patient not found"), err)
}
