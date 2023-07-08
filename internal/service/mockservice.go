package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/ankit/project/message-quening-system/internal/constants"
	"github.com/ankit/project/message-quening-system/internal/db"
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var mockProductClient *MockProductService
var mockMessageChan chan models.Message

type MockProductService struct {
	MockRepo db.MockProductDBService
	writer   MockKafkaWriter
	reader   MockKafkaReader
	Product  *models.Product
	User     *models.User
}

func NewMockProductService(conn db.MockProductDBService, writer *MockKafkaWriter, reader *MockKafkaReader) *MockProductService {
	mockProductClient = &MockProductService{
		MockRepo: conn,
		writer:   *writer,
		reader:   *reader,
		// product:  models.Product{},
		// user:     models.User{},
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

func (m *MockProductService) produceMessages(ctx *gin.Context, messageChan chan models.Message, mockWriter *MockKafkaWriter) error {
	for message := range messageChan {
		// Serialize the message data
		messageData, err := json.Marshal(message)
		if err != nil {
			return err
		}
		utils.Logger.Info("producers marshall the message")
		// Create a Kafka message with the serialized data
		kafkaMessage := kafka.Message{
			Key:   []byte(message.ProductID),
			Value: messageData,
		}
		fmt.Println("2")
		utils.Logger.Info("producers forms the kafka message")
		// Write the Kafka message to the writer
		err = mockWriter.WriteMessages(ctx, kafkaMessage)
		if err != nil {
			return err
		} else {
			utils.Logger.Info("producers writes the kafka message")
			break
		}
	}
	return nil
}

func (m *MockProductService) consumeMessages(ctx *gin.Context, messageChan chan models.Message, mockReader *MockKafkaReader) error {
	for {
		message, err := mockReader.ReadMessage(context.Background())
		if err != nil {
			return fmt.Errorf("failed to read message from Kafka: %v", err)
		}
		utils.Logger.Info("mock consumer has successfully read the message")
		var receivedMessage models.Message
		err = json.Unmarshal(message.Value, &receivedMessage)
		if err != nil {
			return fmt.Errorf("failed to unmarshal message: %v", err)
		}
		utils.Logger.Info("mock consumer has successfully unmarshall the message")
		// Add the received message to the message channel
		messageChan <- receivedMessage

		// Process the received message (e.g., download and compress product images)
		compressedImagesPaths, productErr := m.downloadAndCompressProductImages(ctx, receivedMessage)
		if productErr != nil {
			utils.Logger.Error("Error downloading and compressing images", zap.String("txid", ctx.Request.Header.Get(constants.TransactionID)))
			return fmt.Errorf("error downloading and compressing images: %v", productErr)
		}
		utils.Logger.Info("mock consumer has successfully downloaded and compressed the images")

		// Update the mock database with the compressed_product_images
		productErr = m.updateCompressedProductImages(ctx, fmt.Sprint(receivedMessage.ProductID), compressedImagesPaths)
		if productErr != nil {
			utils.Logger.Error("Error updating compressed images in DB", zap.String("txid", ctx.Request.Header.Get(constants.TransactionID)))
			return fmt.Errorf("error updating compressed images in DB: %v", productErr)
		}
		utils.Logger.Info("mock consumer has successfully updated the mock db with the compressed images path")
		if m.reader.ReadMessageCalled {
			break
		}
	}
	fmt.Println("hello")
	return nil
}

func (m *MockProductService) downloadAndCompressProductImages(ctx *gin.Context, msg models.Message) ([]string, *producterror.ProductError) {
	// Simple mock implementation
	productID, _ := strconv.Atoi(msg.ProductID)
	images, err := m.MockRepo.GetProductImages(ctx, productID)
	if err != nil {
		return []string{}, err
	}

	utils.Logger.Info(fmt.Sprintf("Images returned from mock db are %v", images))

	/*
		Pls, change this according to your directory structure
	*/
	inputPath := "/Users/ankitchahal/dev/go/src/github.com/ankit/project/message-quening-system/cmd/Images/DO-NOT-DELETE1.jpg"
	outputPath := "/Users/ankitchahal/dev/go/src/github.com/ankit/project/message-quening-system/cmd/Images/DO-NOT-DELETE1.jpg"

	utils.Logger.Info("calling image resize functionality to resize image")
	err = m.resizeImage(ctx, inputPath, outputPath, 50, 50)
	if err != nil {
		return []string{}, err
	}
	localImagesPath := []string{"local/image1.jpg", "local/image2.jpg"}
	return localImagesPath, nil
}

func (m *MockProductService) resizeImage(ctx *gin.Context, inputPath, outputPath string, width, height int) *producterror.ProductError {
	// Simple imlementation of resizeImage
	utils.Logger.Info("image resize is successfully done")
	return nil
}

func (m *MockProductService) updateCompressedProductImages(ctx *gin.Context, productID string, compressedImagesPaths []string) *producterror.ProductError {
	pID, _ := strconv.Atoi(productID)

	err := m.MockRepo.UpdateCompressedProductImages(ctx, pID, compressedImagesPaths)
	if err != nil {
		return err
	}
	utils.Logger.Info("update image compressed successfully")
	return nil
}

func (m *MockProductService) addProduct(context *gin.Context, product models.Product) (*int, *producterror.ProductError) {
	productId, err := m.MockRepo.AddProduct(context, product)
	if err != nil {
		return nil, err
	}
	return productId, err
}

// This function needs improvment
func (m *MockProductService) AddProduct(context *gin.Context, product models.Product) error {
	// Mock implementation for creating a product in the database
	var productDetails models.Product
	txid := context.Request.Header.Get(constants.TransactionID)
	utils.Logger.Info("Request received successfully at service layer to add the product", zap.String("txid", txid))
	mockMessageChan = make(chan models.Message, 1)
	//Define the expected message

	expectedMessage := models.Message{
		ProductID: "123",
	}

	//Create a channel for the messages

	mockMessageChan <- expectedMessage
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := m.produceMessages(context, mockMessageChan, &m.writer)
		if err != nil {
			utils.Logger.Error("Error producing messages:", zap.Error(err))
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to produce messages"})
		}
	}()
	wg.Wait()
	wg.Add(1)
	go func() {

		err := m.consumeMessages(context, mockMessageChan, &m.reader)
		if err != nil {
			utils.Logger.Error("Error consuming messages:", zap.Error(err))
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to consume messages"})
		} else {
			wg.Done()
		}
	}()
	fmt.Println("both go routine")
	productID, err := m.addProduct(context, productDetails)
	wg.Wait()
	if err != nil {
		return fmt.Errorf("error while adding product in  mock db %v", err)
	} else {
		utils.Logger.Info(fmt.Sprintf("product id %v  is successfully added.", *productID))
		return nil
	}
	return nil
}

func (m *MockProductService) AddUser(ctx *gin.Context, user models.User) error {
	// Mock implementation for creating a product in the database
	utils.Logger.Info("mock service layer called for adding a user")
	userId, err := m.addUser(ctx, user)
	if err != nil {
		return nil
	}
	utils.Logger.Info(fmt.Sprintf("user Id add in mock db is %v", *userId))
	return nil
}

func (m *MockProductService) addUser(context *gin.Context, user models.User) (*int, *producterror.ProductError) {
	utils.Logger.Info("mock service layer called for adding a user in mock db")
	userId, err := m.MockRepo.AddUser(context, user)
	if err != nil {
		return nil, err
	}
	return userId, err
}

func NewMockKafkaWriter() *MockKafkaWriter {
	return &MockKafkaWriter{}
}
