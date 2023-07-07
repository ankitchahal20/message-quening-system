package db

import (
	"database/sql/driver"
	"log"
	"net/http"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/ankit/project/message-quening-system/internal/constants"
	"github.com/ankit/project/message-quening-system/internal/models"
)

func TestAddProduct(t *testing.T) {
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

	userID := 1001
	productPrice := 10
	var images []string = []string{"image1.jpg", "image2.jpg"}

	productDetails := models.Product{
		UserID:             &userID,
		ProductName:        "Test Product",
		ProductDescription: "This is a test product",
		ProductImages:      images,
		ProductPrice:       &productPrice,
	}

	productID := 123

	// Define the query and expected arguments
	query := `INSERT INTO products(product_name, product_description, product_images, product_price, 
		compressed_product_images, created_at, updated_at, user_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING product_id`
	expectedArgs := []driver.Value{
		productDetails.ProductName,
		productDetails.ProductDescription,
		pq.Array(productDetails.ProductImages),
		productDetails.ProductPrice,
		pq.Array(productDetails.CompressedProductImages),
		productDetails.CreatedAt,
		productDetails.UpdatedAt,
		productDetails.UserID,
	}

	// Define the expected result from the database
	rows := sqlmock.NewRows([]string{"product_id"}).AddRow(&productID)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(expectedArgs...).WillReturnRows(rows)

	// Call the AddProduct function
	_, productErr := p.AddProduct(ctx, productDetails)
	assert.Nil(t, productErr)

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
