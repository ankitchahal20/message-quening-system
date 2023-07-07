package service

import (
	"log"

	"github.com/ankit/project/message-quening-system/internal/config"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pelletier/go-toml"
)

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
// 	postgres, err := db.New()
// 	if err != nil {
// 		log.Fatal("Unable to connect to DB : ", err)
// 	}

// 	kafkaWriter := kafka.IntializeKafkaProducerWriter()
// 	defer kafkaWriter.Close()
// 	kafkaReader := kafka.IntializeKafkaConsumerReader()
// 	defer kafkaReader.Close()

// 	utils.InitLogClient()
// 	_ = NewProductService(postgres, kafkaWriter, kafkaReader)

// 	// Create a test product
// 	productPrice := 2
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

// 	jsonValue, _ := json.Marshal(product)

// 	req, _ := http.NewRequest(http.MethodPost, "/v1/product/create", bytes.NewBuffer(jsonValue))
// 	req.Header.Add(constants.ContentType, "application/json")
// 	w := httptest.NewRecorder()
// 	_, e := gin.CreateTestContext(w)
// 	e.Use(AddProduct())
// 	e.ServeHTTP(w, req)
// 	assert.Equal(t, http.StatusOK, w.Code)

// }
