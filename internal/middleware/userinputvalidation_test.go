package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ankit/project/message-quening-system/internal/config"
	"github.com/ankit/project/message-quening-system/internal/constants"
	"github.com/ankit/project/message-quening-system/internal/models"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestValidateUserRequestInput(t *testing.T) {
	config.InitGlobalConfig()

	// init logging client
	utils.InitLogClient()

	userID := 10
	latitude := 37.1234
	longitude := -122.5678
	// Case 1 : user name is  ,issing
	requestFields := models.User{
		//Name:      "Zocket",
		ID:        &userID,
		Mobile:    "12344566",
		Latitude:  &latitude,
		Longitude: &longitude,
	}

	jsonValue, _ := json.Marshal(requestFields)

	w := httptest.NewRecorder()
	_, e := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/v1/productapi/user/create", bytes.NewBuffer(jsonValue))
	req.Header.Add(constants.ContentType, "application/json")
	e.Use(ValidateProductInputRequest())
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Case 2 : Product mobile number is missing
	requestFields = models.User{
		Name: "Zocket",
		ID:   &userID,
		//Mobile:    "12344566",
		Latitude:  &latitude,
		Longitude: &longitude,
	}

	jsonValue, _ = json.Marshal(requestFields)

	w = httptest.NewRecorder()
	_, e = gin.CreateTestContext(w)
	req, _ = http.NewRequest(http.MethodPost, "/v1/productapi/user/create", bytes.NewBuffer(jsonValue))
	req.Header.Add(constants.ContentType, "application/json")
	e.Use(ValidateProductInputRequest())
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Case 3 : user latitude location missing
	requestFields = models.User{
		Name:   "Zocket",
		ID:     &userID,
		Mobile: "12344566",
		//Latitude:  &latitude,
		Longitude: &longitude,
	}

	jsonValue, _ = json.Marshal(requestFields)

	w = httptest.NewRecorder()
	_, e = gin.CreateTestContext(w)
	req, _ = http.NewRequest(http.MethodPost, "/v1/productapi/user/create", bytes.NewBuffer(jsonValue))
	req.Header.Add(constants.ContentType, "application/json")
	e.Use(ValidateProductInputRequest())
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Case 4 : user longitude location missing
	requestFields = models.User{
		Name:     "Zocket",
		ID:       &userID,
		Mobile:   "12344566",
		Latitude: &latitude,
		//Longitude: &longitude,
	}

	jsonValue, _ = json.Marshal(requestFields)

	w = httptest.NewRecorder()
	_, e = gin.CreateTestContext(w)
	req, _ = http.NewRequest(http.MethodPost, "/v1/productapi/user/create", bytes.NewBuffer(jsonValue))
	req.Header.Add(constants.ContentType, "application/json")
	e.Use(ValidateProductInputRequest())
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

}
