package integrationtest_test

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/ankit/project/message-quening-system/internal/config"
	"github.com/ankit/project/message-quening-system/internal/constants"
	"github.com/ankit/project/message-quening-system/internal/db"
	"github.com/ankit/project/message-quening-system/internal/models"
	"github.com/ankit/project/message-quening-system/internal/service"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
)

var messageChan chan models.Message

func SetUpRouter() *gin.Engine {
	router := gin.New()
	return router
}

var isTestRunning = false

func initGlobalConfig() error {
	if !isTestRunning {
		configTree, err := toml.LoadFile("./../config/default.toml")
		if err != nil {
			log.Printf("Error while loading deafault.toml file : %v ", err)
			return err
		}

		var appConfig config.GlobalConfig
		err = configTree.Unmarshal(&appConfig)
		if err != nil {
			log.Printf("Error while unmarshalling config : %v", err)
			return err
		}

		config.SetConfig(appConfig)
		isTestRunning = true
		return nil
	}
	return nil
}

func TestIntegration_AddProduct(t *testing.T) {
	initGlobalConfig()

	utils.InitLogClient()
	// _ = NewProductService(postgres, kafkaWriter, kafkaReader)

	productPrice := 6
	userId := 2
	product := models.Product{
		UserID:             &userId,
		ProductName:        "Test Product",
		ProductDescription: "Test Description",
		ProductImages:      []string{"https://cdn.pixabay.com/photo/2013/10/15/09/12/flower-195893_150.jpg", "https://images.pexels.com/photos/2014422/pexels-photo-2014422.jpeg", "https://images.pexels.com/photos/2014421/pexels-photo-2014421.jpeg"},
		ProductPrice:       &productPrice,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	mp := &db.MockPostgres{
		Product: &product,
	}

	mockProductService := service.NewMockProductService(mp, &service.MockKafkaWriter{}, &service.MockKafkaReader{})
	mockProductService.Product = &product

	transactionID := uuid.New().String()
	ctx := &gin.Context{
		Request: &http.Request{
			Header: http.Header{
				constants.TransactionID: []string{transactionID},
			}},
	}
	ctx.Request.Header.Set(constants.TransactionID, transactionID)
	err := mockProductService.AddProduct(ctx, product)
	assert.Nil(t, err)

}

func TestIntegration_AddUser(t *testing.T) {
	initGlobalConfig()

	utils.InitLogClient()
	// _ = NewProductService(postgres, kafkaWriter, kafkaReader)

	latitude := 12.34
	longitude := 56.78
	user := models.User{
		Name:      "Test User",
		Mobile:    "1234567890",
		Latitude:  &latitude,
		Longitude: &longitude,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mp := &db.MockPostgres{
		User: &user,
	}

	mockProductService := service.NewMockProductService(mp, &service.MockKafkaWriter{}, &service.MockKafkaReader{})
	mockProductService.User = &user

	transactionID := uuid.New().String()
	ctx := &gin.Context{
		Request: &http.Request{
			Header: http.Header{
				constants.TransactionID: []string{transactionID},
			}},
	}
	ctx.Request.Header.Set(constants.TransactionID, transactionID)
	err := mockProductService.AddUser(ctx, user)
	assert.Nil(t, err)
}
