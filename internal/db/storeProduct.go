package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ankit/project/message-quening-system/internal/constants"
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type postgres struct{ db *sql.DB }

type ProductService interface {
	CreateProduct(*gin.Context, models.Product) *producterror.ProductError
	GetProductImages(*gin.Context, int) ([]string, *producterror.ProductError)
	UpdateCompressedProductImages(*gin.Context, int, []string) *producterror.ProductError
}

var (
	ErrNoRowFound         = errors.New("no row found in DB for the given product id")
	ErrUnableToInsertARow = errors.New("unable to perform insert opertion on the products table")
	ErrUnableToSelectRows = errors.New("unable to perform select opertion on the products table")
	ErrScanningRows       = errors.New("unable to scan rows")
	ErrZeroRowsFound      = errors.New("no row found in DB for the given product id")
)

func (p postgres) CreateProduct(ctx *gin.Context, productDetails models.Product) *producterror.ProductError {
	query := `INSERT INTO products(product_id, product_name, product_description, product_images, product_price, 
		compressed_product_images, created_at, updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`

	// PostgreSQL driver for Go does not support passing slices directly as arguments to SQL queries.
	// So, convert the slice of strings into a supported data type for the query.
	productImagesArray := pq.Array(productDetails.ProductImages)
	compressedImagesArray := pq.Array(productDetails.CompressedProductImages)

	_, err := p.db.Exec(query, productDetails.ProductID, productDetails.ProductName, productDetails.ProductDescription, productImagesArray, productDetails.ProductPrice,
		compressedImagesArray, productDetails.CreatedAt, productDetails.UpdatedAt)
	if err != nil {
		log.Println("unable to insert product details info in table : ", err, "/n", err.Error())
		if strings.Contains(err.Error(), "duplicate key value") {
			fmt.Println("Hello ")
			return &producterror.ProductError{
				Trace:   ctx.Request.Header.Get(constants.TransactionID),
				Code:    http.StatusBadRequest,
				Message: "product already added",
			}
		} else {
			return &producterror.ProductError{
				Trace:   ctx.Request.Header.Get(constants.TransactionID),
				Code:    http.StatusInternalServerError,
				Message: "unable to add product details",
			}
		}
	}
	return nil
}

func (p postgres) GetProductImages(ctx *gin.Context, productID int) ([]string, *producterror.ProductError) {
	query := `SELECT product_images FROM products WHERE product_id=$1`
	var images []string
	if err := p.db.QueryRow(query, productID).Scan(pq.Array(&images)); err != nil {
		return images, &producterror.ProductError{
			Code:    http.StatusInternalServerError,
			Message: "Unable to get product images from DB",
			Trace:   ctx.Request.Header.Get(constants.TransactionID),
		}
	}
	return images, nil

}

func (p postgres) UpdateCompressedProductImages(ctx *gin.Context, productID int, compressedImages []string) *producterror.ProductError {
	query := "UPDATE products SET compressed_product_images = $1 WHERE product_id = $2"
	compressedImagesArray := pq.Array(compressedImages)

	_, err := p.db.Exec(query, compressedImagesArray, productID)
	if err != nil {
		return &producterror.ProductError{
			Code:    http.StatusInternalServerError,
			Message: "Unable to add compressed images in DB",
			Trace:   ctx.Request.Header.Get(constants.TransactionID),
		}
	}

	return nil
}
