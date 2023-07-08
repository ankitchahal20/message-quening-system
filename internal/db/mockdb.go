package db

import (
	"fmt"
	"time"

	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
)

type MockProductDBService interface {
	// product
	AddProduct(*gin.Context, models.Product) (*int, *producterror.ProductError)
	GetProductImages(*gin.Context, int) ([]string, *producterror.ProductError)
	UpdateCompressedProductImages(*gin.Context, int, []string) *producterror.ProductError

	// user
	AddUser(*gin.Context, models.User) (*int, *producterror.ProductError)
}

type MockPostgres struct {
	Product *models.Product
}

func (m *MockPostgres) AddProduct(ctx *gin.Context, product models.Product) (*int, *producterror.ProductError) {
	m.Product.ProductName = product.ProductName
	m.Product.CreatedAt = product.CreatedAt
	m.Product.CompressedProductImages = product.CompressedProductImages
	m.Product.ProductImages = product.ProductImages
	m.Product.ProductDescription = product.ProductDescription
	m.Product.UpdatedAt = product.UpdatedAt
	productId := 101
	return &productId, nil
}

func (m *MockPostgres) GetProductImages(*gin.Context, int) ([]string, *producterror.ProductError) {
	return m.Product.ProductImages, nil
}

func (m *MockPostgres) UpdateCompressedProductImages(ctx *gin.Context, productID int, compressedImagesPaths []string) *producterror.ProductError {
	utils.Logger.Info("Compressed images are stored in mock db successfully")
	fmt.Println("compressedImages : ", compressedImagesPaths)
	m.Product.UpdatedAt = time.Now().UTC()
	m.Product.CompressedProductImages = append(m.Product.CompressedProductImages, compressedImagesPaths...)
	return nil
}

func (m *MockPostgres) AddUser(ctx *gin.Context, user models.User) (*int, *producterror.ProductError) {
	fmt.Println("user added successfully ")
	userId := 1
	return &userId, nil
}
