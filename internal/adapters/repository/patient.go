package repository

import (
	"errors"

	"github.com/datphamcode295/go-lambda-pulumi/internal/core/domain"
)

func (u *DB) GetPatient(id string) (*domain.Patient, error) {
	patient := &domain.Patient{}

	req := u.db.First(&patient, "id = ? ", id)
	if req.RowsAffected == 0 {
		return nil, errors.New("patient not found")
	}

	return patient, nil
}
