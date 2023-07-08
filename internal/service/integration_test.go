package service_test

import (
	"log"

	"github.com/ankit/project/message-quening-system/internal/config"
	"github.com/ankit/project/message-quening-system/internal/models"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pelletier/go-toml"
)

var messageChan chan models.Message

func SetUpRouter() *gin.Engine {
	router := gin.New()
	return router
}

var isTestRunning = false

func initGlobalConfig() error {
	if !isTestRunning {
		configTree, err := toml.LoadFile("./../../config/default.toml")
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

// func TestIntegration_AddProduct(t *testing.T) {
// 	initGlobalConfig()
// 	// postgres, err := db.New()
// 	// if err != nil {
// 	// 	log.Fatal("Unable to connect to DB : ", err)
// 	// }

// 	// kafkaWriter := kafka.IntializeKafkaProducerWriter()
// 	// defer kafkaWriter.Close()
// 	// kafkaReader := kafka.IntializeKafkaConsumerReader()
// 	// defer kafkaReader.Close()

// 	utils.InitLogClient()
// 	// _ = NewProductService(postgres, kafkaWriter, kafkaReader)

// 	productPrice := 6
// 	userId := 2
// 	product := models.Product{
// 		UserID:             &userId,
// 		ProductName:        "Test Product",
// 		ProductDescription: "Test Description",
// 		ProductImages:      []string{"https://cdn.pixabay.com/photo/2013/10/15/09/12/flower-195893_150.jpg", "https://images.pexels.com/photos/2014422/pexels-photo-2014422.jpeg", "https://images.pexels.com/photos/2014421/pexels-photo-2014421.jpeg"},
// 		ProductPrice:       &productPrice,
// 		CreatedAt:          time.Now(),
// 		UpdatedAt:          time.Now(),
// 	}
// 	mp := &db.MockPostgres{
// 		Product: &product,
// 	}

// 	mockProductService := service.NewMockProductService(mp, &service.MockKafkaWriter{}, &service.MockKafkaReader{})
// 	mockProductService.Product = &product

// 	// Create a test product

// 	//jsonValue, _ := json.Marshal(product)
// 	transactionID := uuid.New().String()
// 	ctx := &gin.Context{
// 		Request: &http.Request{
// 			Header: http.Header{
// 				constants.TransactionID: []string{transactionID},
// 			}},
// 	}
// 	// req, _ := http.NewRequest(http.MethodPost, "/v1/productapi/product/create", bytes.NewBuffer(jsonValue))
// 	// req.Header.Add(constants.ContentType, "application/json")
// 	// w := httptest.NewRecorder()
// 	// ctx, e := gin.CreateTestContext(w)
// 	// ctx.Request = req
// 	ctx.Request.Header.Set(constants.TransactionID, transactionID)
// 	// e := gin.New()
// 	err := mockProductService.AddProduct(ctx, product)
// 	//e.ServeHTTP(w, req)
// 	assert.Nil(t, err)

// }
