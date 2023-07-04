package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ankit/project/message-quening-system/cmd/consumer"
	"github.com/ankit/project/message-quening-system/cmd/producer"
	"github.com/ankit/project/message-quening-system/internal/db"
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/segmentio/kafka-go"
)

var productClient *ProductService

type ProductService struct {
	repo db.ProductService
}

func NewProductService(conn db.ProductService) {
	productClient = &ProductService{
		repo: conn,
	}
}

func CreateProduct() func(ctx *gin.Context) {
	return func(context *gin.Context) {
		var productDetails models.Product
		if err := context.ShouldBindBodyWith(&productDetails, binding.JSON); err == nil {

			product, err := productClient.createProduct(context, productDetails)
			if err != nil {
				context.Writer.WriteHeader(http.StatusInternalServerError)
			} else {
				json.NewEncoder(context.Writer).Encode(map[string]string{"product_name": product.ProductName})
				context.JSON(http.StatusCreated, product)
			}
		} else {
			context.JSON(http.StatusBadRequest, gin.H{"Unable to marshal the request body": err.Error()})
		}
	}
}

func (service *ProductService) createProduct(ctx *gin.Context, productDetails models.Product) (models.Product, *producterror.ProductError) {
	fmt.Println("hello from product service", productDetails)
	productDetails.CreatedAt = time.Now().UTC()
	productDetails.UpdatedAt = time.Now().UTC()
	err := service.repo.CreateProduct(ctx, productDetails)
	if err != nil {
		return models.Product{}, err
	}

	fmt.Println("productDetails.ProductID : ", *productDetails.ProductID)

	// Send the product ID to the message channel
	message := models.Message{
		ProductID: fmt.Sprint(*productDetails.ProductID),
		Product:   productDetails,
	}
	utils.MessageChan <- message

	return productDetails, nil
}

func ProduceMessages(messageChan <-chan models.Message) {
	for message := range messageChan {
		// Serialize the message data
		messageData, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling message: %v\n", err)
			continue
		}

		// Create a Kafka message with the serialized data
		kafkaMessage := kafka.Message{
			Key:   []byte(message.ProductID),
			Value: messageData,
		}

		// Write the message to the Kafka topic
		err = producer.KafkaWriter.WriteMessages(context.Background(), kafkaMessage)
		if err != nil {
			log.Printf("Error writing message to Kafka: %v\n", err)
		}
	}
}

func ConsumeMessages(messageChan chan<- models.Message) {
	for {
		// Read the next message from the Kafka topic
		message, err := consumer.KafkaReader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message from Kafka: %v\n", err)
			continue
		}

		// Deserialize the kafka message data into a Message struct
		var receivedMessage models.Message
		err = json.Unmarshal(message.Value, &receivedMessage)
		if err != nil {
			log.Printf("Error unmarshaling message: %v\n", err)
			continue
		}

		// Process the received message (e.g., download and compress product images)

		// Add the compressed image path to the product in the database
		compressedImages := processProductImages(receivedMessage.ProductID)

		// 	Update the database with the compressed_product_images
		updateProductImages(receivedMessage.ProductID, compressedImages)

		// Print the processed message
		log.Printf("Processed message: ProductID=%s, ProductName=%s\n", receivedMessage.ProductID, receivedMessage.Product.ProductName)
	}
}

func processProductImages(productID string) []string {
	// Simulate image compression process.
	fmt.Printf("Product images compressed for product_id: %s\n", productID)

	// Return the compressed image file paths.
	return []string{
		"/path/to/compressed/image1.jpg",
		"/path/to/compressed/image2.jpg",
	}
}

func updateProductImages(productID string, compressedImages []string) {
	// Update the compressed_product_images column in the database (DB implementation not provided).
	fmt.Printf("Compressed images updated for product_id: %s\n", productID)
}
