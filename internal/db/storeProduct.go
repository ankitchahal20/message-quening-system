package db

import (
	"database/sql"
	"errors"

	"github.com/ankit/project/message-quening-system/internal/models"
)

type postgres struct{ db *sql.DB }

type ProductService interface {
	CreateProduct(models.Product) error
}

var (
	ErrNoRowFound         = errors.New("no row found in DB for the given short url")
	ErrUnableToInsertARow = errors.New("unable to perform select opertion on the url table")
	ErrUnableToSelectRows = errors.New("unable to perform select opertion on the url table")
	ErrScanningRows       = errors.New("unable to scan rows")
	ErrZeroRowsFound      = errors.New("no row found in DB for the given short url")
)

func (p postgres) CreateProduct(urlInfo models.Product) error {
	return nil
}
