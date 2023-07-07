package db

import (
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/gin-gonic/gin"
)

type MockProductDBService interface {
	AddProduct(*gin.Context, models.Product) *producterror.ProductError
	GetProductImages(*gin.Context, int) ([]string, *producterror.ProductError)
	UpdateCompressedProductImages(*gin.Context, int, []string) *producterror.ProductError
}

type MockPostgres struct{}

func (m MockPostgres) AddProduct(*gin.Context, models.Product) *producterror.ProductError {
	return nil
}

func (m MockPostgres) GetProductImages(*gin.Context, int) ([]string, *producterror.ProductError) {
	return []string{}, nil
}

func (m MockPostgres) UpdateCompressedProductImages(*gin.Context, int, []string) *producterror.ProductError {
	return nil
}
