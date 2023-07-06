package db

import (
	"database/sql"
	"log"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/ankit/project/message-quening-system/internal/constants"
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
)

func TestCreateProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("error creating mock database: %v", err)
	}
	defer db.Close()

	p := postgres{
		db: db,
	}

	transactionID := uuid.New().String()
	ctx := &gin.Context{
		Request: &http.Request{
			Header: http.Header{
				constants.TransactionID: []string{transactionID},
			}},
	}

	ctx.Request.Header.Set(constants.TransactionID, transactionID)

	productID := 1001
	productPrice := 10
	var images []string = []string{"image1.jpg", "image2.jpg"}
	productImages := strings.Join(images, ", ")
	product := models.Product{
		ProductID:          &productID,
		ProductName:        "Test Product",
		ProductDescription: "This is a test product",
		ProductImages:      []string{productImages},
		ProductPrice:       &productPrice,
	}

	t.Run("Successfully create product", func(t *testing.T) {
		// Set up the expected SQL query
		mock.ExpectExec("INSERT INTO products").
			WillReturnResult(sqlmock.NewResult(1, 1))

		product.CreatedAt = time.Now().UTC()
		product.UpdatedAt = time.Now().UTC()
		// Invoking the function being tested
		err := p.CreateProduct(ctx, product)

		assert.Nil(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error: Unable to add product details", func(t *testing.T) {
		// Set up the expected SQL query
		mock.ExpectExec("INSERT INTO products").
			WillReturnError(sql.ErrConnDone)

		// Invoking the function being tested
		err := p.CreateProduct(ctx, product)

		expectedErr := &producterror.ProductError{
			Message: "unable to add product details",
			Code:    http.StatusInternalServerError,
		}

		assert.Equal(t, expectedErr.Code, err.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetProductImages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("error creating mock database: %v", err)
	}
	defer db.Close()

	p := postgres{
		db: db,
	}

	productID := 1
	images := []string{"image1.jpg", "image2.jpg"}

	// Adding rows
	rows := sqlmock.NewRows([]string{"product_images"}).
		AddRow(pq.Array(images))

	// Set up the expected SQL query
	mock.ExpectQuery("SELECT product_images FROM products").
		WithArgs(productID).
		WillReturnRows(rows)

	ctx := &gin.Context{}
	// Invoking the function being tested
	retrievedImages, productErr := p.GetProductImages(ctx, productID)

	assert.Nil(t, productErr)
	assert.Equal(t, images, retrievedImages)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCompressedProductImages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("error creating mock database: %v", err)
	}
	defer db.Close()

	p := postgres{
		db: db,
	}

	transactionID := uuid.New().String()
	ctx := &gin.Context{
		Request: &http.Request{
			Header: http.Header{
				constants.TransactionID: []string{transactionID},
			}},
	}

	ctx.Request.Header.Set(constants.TransactionID, transactionID)

	productID := 1
	compressedImages := []string{"image1_compressed.jpg", "image2_compressed.jpg"}

	// Setting up the expected SQL query and result
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE products SET compressed_product_images = $1 WHERE product_id = $2`)).
		WithArgs(pq.Array(compressedImages), productID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Invoking the function being tested
	productErr := p.UpdateCompressedProductImages(ctx, productID, compressedImages)

	// Assert that the returned error is nil
	assert.Nil(t, productErr)

	// Assert that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
