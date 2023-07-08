package db

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ankit/project/message-quening-system/internal/constants"
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

var (
	ErrNoRowFound         = errors.New("no row found in DB for the given product id")
	ErrUnableToInsertARow = errors.New("unable to perform insert opertion on the products table")
	ErrUnableToSelectRows = errors.New("unable to perform select opertion on the products table")
	ErrScanningRows       = errors.New("unable to scan rows")
	ErrZeroRowsFound      = errors.New("no row found in DB for the given product id")
)

func (p postgres) AddProduct(ctx *gin.Context, productDetails models.Product) (*int, *producterror.ProductError) {
	query := `INSERT INTO products(product_name, product_description, product_images, product_price, 
		compressed_product_images, created_at, updated_at, user_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING product_id`

	// PostgreSQL driver for Go does not support passing slices directly as arguments to SQL queries.
	// So, convert the slice of strings into a supported data type for the query.
	productImagesArray := pq.Array(productDetails.ProductImages)
	compressedImagesArray := pq.Array(productDetails.CompressedProductImages)

	productID := 0
	err := p.db.QueryRow(query, productDetails.ProductName, productDetails.ProductDescription, productImagesArray, productDetails.ProductPrice,
		compressedImagesArray, productDetails.CreatedAt, productDetails.UpdatedAt, productDetails.UserID).Scan(&productID)
	if err != nil {
		log.Println("unable to insert product details info in table : ", err, "/n", err.Error())
		if strings.Contains(err.Error(), "duplicate key value") {
			return nil, &producterror.ProductError{
				Trace:   ctx.Request.Header.Get(constants.TransactionID),
				Code:    http.StatusBadRequest,
				Message: "product already added",
			}
		} else if strings.Contains(err.Error(), "violates foreign key constraint") {
			return nil, &producterror.ProductError{
				Trace:   ctx.Request.Header.Get(constants.TransactionID),
				Code:    http.StatusBadRequest,
				Message: "user id is not found",
			}
		} else {
			return nil, &producterror.ProductError{
				Trace:   ctx.Request.Header.Get(constants.TransactionID),
				Code:    http.StatusInternalServerError,
				Message: "unable to add product details",
			}
		}
	}
	utils.Logger.Info("added product in db successfully")

	return &productID, nil
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
	query := "UPDATE products SET compressed_product_images = $1, updated_at=$2 WHERE product_id = $3"
	compressedImagesArray := pq.Array(compressedImages)

	_, err := p.db.Exec(query, compressedImagesArray, time.Now().UTC(), productID)
	if err != nil {
		return &producterror.ProductError{
			Code:    http.StatusInternalServerError,
			Message: "Unable to add compressed images in DB",
			Trace:   ctx.Request.Header.Get(constants.TransactionID),
		}
	}

	return nil
}
