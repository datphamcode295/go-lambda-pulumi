package repository

import (
	"errors"
	"fmt"

	"github.com/datphamcode295/go-lambda-pulumi/internal/core/domain"
)

func (u *DB) CreateTransaction(transaction domain.Transaction) (*domain.Transaction, error) {
	fmt.Println("Creating transaction", transaction)
	req := u.db.Create(&transaction)
	if req.RowsAffected == 0 {
		return nil, errors.New("transaction not created")
	}

	return &transaction, nil
}
