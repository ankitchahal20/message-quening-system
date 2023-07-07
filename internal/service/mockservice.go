package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ankit/project/message-quening-system/internal/db"
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var mockProductClient *MockProductService

type MockProductService struct {
	MockRepo db.MockProductDBService
	writer   MockKafkaWriter
	reader   MockKafkaReader
}

func NewMockProductService(conn db.MockProductDBService, writer *MockKafkaWriter, reader *MockKafkaReader) *MockProductService {
	mockProductClient = &MockProductService{
		MockRepo: conn,
		writer:   *writer,
		reader:   *reader,
	}
	return mockProductClient
}

type MockKafkaWriter struct {
	Messages []kafka.Message
}

type MockKafkaReader struct {
	ReadMessageCalled bool
	ReceivedMessage   kafka.Message
}

func (m *MockKafkaReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	m.ReadMessageCalled = true

	messageData := []byte(`{"product_id": "123", "product": {"product_name": "Test Product"}}`)
	message := kafka.Message{
		Key:   []byte("mock-key"),
		Value: messageData,
	}

	m.ReceivedMessage = message

	return message, nil
}

func (w *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	w.Messages = append(w.Messages, msgs...)
	return nil
}

func (m *MockProductService) ProduceMessages(ctx *gin.Context, messageChan <-chan models.Message, mockWriter *MockKafkaWriter) error {
	for message := range messageChan {
		fmt.Println("Producing message:", message)

		// Serialize the message data
		messageData, err := json.Marshal(message)
		if err != nil {
			return err
		}

		// Create a Kafka message with the serialized data
		kafkaMessage := kafka.Message{
			Key:   []byte(message.ProductID),
			Value: messageData,
		}

		// Write the Kafka message to the writer
		err = mockWriter.WriteMessages(ctx, kafkaMessage)
		fmt.Println("WERR : ", err)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MockProductService) ConsumeMessages(ctx *gin.Context, messageChan chan models.Message, mockReader *MockKafkaReader) error {
	for {
		message, err := mockReader.ReadMessage(context.Background())
		if err != nil {
			return fmt.Errorf("failed to read message from Kafka: %v", err)
		}
		var receivedMessage models.Message
		err = json.Unmarshal(message.Value, &receivedMessage)
		if err != nil {
			return fmt.Errorf("failed to unmarshal message: %v", err)
		}

		// Add the received message to the message channel
		messageChan <- receivedMessage

		// Process the received message (e.g., download and compress product images)
		compressedImages, productErr := m.DownloadAndCompressProductImages(ctx, receivedMessage)
		if productErr != nil {
			utils.Logger.Error("Error downloading and compressing images:")
			return fmt.Errorf("error downloading and compressing images: %v", productErr)
		}

		// Update the database with the compressed_product_images
		err = m.UpdateCompressedProductImages(ctx, fmt.Sprint(receivedMessage.Product.ProductID), compressedImages)
		if err != nil {
			utils.Logger.Error("Error updating compressed images in DB:", zap.Error(err))
			return fmt.Errorf("error updating compressed images in DB: %w", err)
		}
	}
}

func (m *MockProductService) DownloadAndCompressProductImages(ctx *gin.Context, msg models.Message) ([]string, *producterror.ProductError) {
	return []string{"image1", "image2"}, nil
}

func (m *MockProductService) UpdateCompressedProductImages(ctx *gin.Context, productID string, compressedImages []string) error {
	fmt.Println("Image compressed successfully")
	return nil
}

func (m *MockProductService) CreateProduct(ctx context.Context, product models.Product) error {
	// Mock implementation for creating a product in the database
	log.Println("Creating product:", product)
	return nil
}

func (m *MockProductService) GetProduct(ctx context.Context, productID int) (models.Product, error) {
	// Mock implementation for retrieving a product from the database
	log.Println("Getting product:", productID)
	return models.Product{}, nil
}

func (m *MockProductService) UpdateProduct(ctx context.Context, productID int, updatedProduct models.Product) error {
	// Mock implementation for updating a product in the database
	log.Println("Updating product:", productID)
	return nil
}

func NewMockKafkaWriter() *MockKafkaWriter {
	return &MockKafkaWriter{}
}
