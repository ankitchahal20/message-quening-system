package db

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ankit/project/message-quening-system/internal/models"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
)

func TestAddUser(t *testing.T) {
	// Create a new mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Create a new instance of the Postgres struct with the mock database
	p := postgres{db: db}
	utils.InitLogClient()

	// Set up the test input data
	latitude := 12.34
	longitude := 56.78
	userDetails := models.User{
		Name:      "Ankit Chahal",
		Mobile:    "1234567890",
		Latitude:  &latitude,
		Longitude: &longitude,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set up the expected SQL query and result
	mock.ExpectQuery(`INSERT INTO users`).WithArgs(
		userDetails.Name,
		userDetails.Mobile,
		userDetails.Latitude,
		userDetails.Longitude,
		userDetails.CreatedAt,
		userDetails.UpdatedAt,
	).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Call the AddUser function with the mock context and user details
	ctx := &gin.Context{}
	userID, productErr := p.AddUser(ctx, userDetails)

	// Check the returned values
	if productErr != nil {
		t.Errorf("AddUser returned an unexpected error: %v", err)
	}
	if userID == nil || *userID != 1 {
		t.Errorf("AddUser returned an unexpected user ID: %v", userID)
	}

	// Check that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}
