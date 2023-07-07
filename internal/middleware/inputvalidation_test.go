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

func TestValidateRequestInput(t *testing.T) {
	config.InitGlobalConfig()

	// init logging client
	utils.InitLogClient()

	productPrice := 10
	// Case 1 : User ID Missing
	requestFields := models.Product{
		ProductName:        "Zocket",
		ProductDescription: "some-random-description",
		ProductImages:      []string{"https://cdn.pixabay.com/photo/2013/10/15/09/12/flower-195893_150.jpg"},
		ProductPrice:       &productPrice,
	}

	jsonValue, _ := json.Marshal(requestFields)

	w := httptest.NewRecorder()
	_, e := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/v1/product/create", bytes.NewBuffer(jsonValue))
	req.Header.Add(constants.ContentType, "application/json")
	e.Use(ValidateInputRequest())
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Case 2 : Product Name Missing
	userId := 1
	requestFields = models.Product{
		UserID:             &userId,
		ProductDescription: "some-random-description",
		ProductImages:      []string{"https://cdn.pixabay.com/photo/2013/10/15/09/12/flower-195893_150.jpg"},
		ProductPrice:       &productPrice,
	}

	jsonValue, _ = json.Marshal(requestFields)

	w = httptest.NewRecorder()
	_, e = gin.CreateTestContext(w)
	req, _ = http.NewRequest(http.MethodPost, "/v1/product/create", bytes.NewBuffer(jsonValue))
	req.Header.Add(constants.ContentType, "application/json")
	e.Use(ValidateInputRequest())
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Case 3 : Product Description Missing
	requestFields = models.Product{
		UserID:        &userId,
		ProductName:   "Zocket",
		ProductImages: []string{"https://cdn.pixabay.com/photo/2013/10/15/09/12/flower-195893_150.jpg"},
		ProductPrice:  &productPrice,
	}

	jsonValue, _ = json.Marshal(requestFields)

	w = httptest.NewRecorder()
	_, e = gin.CreateTestContext(w)
	req, _ = http.NewRequest(http.MethodPost, "/v1/product/create", bytes.NewBuffer(jsonValue))
	req.Header.Add(constants.ContentType, "application/json")
	e.Use(ValidateInputRequest())
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Case 4 : Product Images Missing
	requestFields = models.Product{
		UserID:       &userId,
		ProductName:  "Zocket",
		ProductPrice: &productPrice,
	}

	jsonValue, _ = json.Marshal(requestFields)

	w = httptest.NewRecorder()
	_, e = gin.CreateTestContext(w)
	req, _ = http.NewRequest(http.MethodPost, "/v1/product/create", bytes.NewBuffer(jsonValue))
	req.Header.Add(constants.ContentType, "application/json")
	e.Use(ValidateInputRequest())
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Case 5 : Product Price Missing
	requestFields = models.Product{
		UserID:        &userId,
		ProductName:   "Zocket",
		ProductImages: []string{"https://cdn.pixabay.com/photo/2013/10/15/09/12/flower-195893_150.jpg"},
	}

	jsonValue, _ = json.Marshal(requestFields)

	w = httptest.NewRecorder()
	_, e = gin.CreateTestContext(w)
	req, _ = http.NewRequest(http.MethodPost, "/v1/product/create", bytes.NewBuffer(jsonValue))
	req.Header.Add(constants.ContentType, "application/json")
	e.Use(ValidateInputRequest())
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

}
