package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ankit/project/message-quening-system/internal/constants"
	"github.com/ankit/project/message-quening-system/internal/db"
	"github.com/ankit/project/message-quening-system/internal/models"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestProduceMessages(t *testing.T) {
	mockWriter := &MockKafkaWriter{}
	productClient := NewMockProductService(&db.MockPostgres{}, mockWriter, &MockKafkaReader{})

	// Define the expected message
	expectedMessage := models.Message{
		ProductID: "123",
		Product: models.Product{
			ProductName: "Test Product",
		},
	}

	// Create a channel for the messages
	transactionID := uuid.New().String()
	messageChan := make(chan models.Message)
	ctx := &gin.Context{
		Request: &http.Request{
			Header: http.Header{
				constants.TransactionID: []string{transactionID},
			}},
	}
	var wg sync.WaitGroup
	wg.Add(1)

	// Start the goroutine for producing messages
	go func() {
		defer wg.Done()
		err := productClient.ProduceMessages(ctx, messageChan, mockWriter)
		if err != nil {
			t.Errorf("Failed to produce messages: %v", err)
		}
	}()
	// Send the expected message to the channel
	messageChan <- expectedMessage

	// Close the channel to indicate that no more messages will be sent
	close(messageChan)

	// Wait for the goroutine to finish
	wg.Wait()

	if len(mockWriter.Messages) != 1 {
		t.Errorf("Unexpected number of messages written. Expected 1, got %d", len(mockWriter.Messages))
	}

	writtenMessage := mockWriter.Messages[0]
	expectedKey := []byte(expectedMessage.ProductID)
	if !bytes.Equal(writtenMessage.Key, expectedKey) {
		t.Errorf("Unexpected message key. Expected %s, got %s", expectedMessage.ProductID, string(writtenMessage.Key))
	}

	// Deserialize the written message value
	var receivedMessage models.Message
	err := json.Unmarshal(writtenMessage.Value, &receivedMessage)
	if err != nil {
		t.Errorf("Failed to unmarshal written message: %v", err)
	}

	if receivedMessage.ProductID != expectedMessage.ProductID {
		t.Errorf("Unexpected message value. Expected %s, got %s", expectedMessage.ProductID, receivedMessage.ProductID)
	}
}

func TestConsumeMessages(t *testing.T) {
	mockReader := &MockKafkaReader{}
	mockProductService := NewMockProductService(&db.MockPostgres{}, &MockKafkaWriter{}, mockReader)

	// Define the expected message
	expectedMessage := models.Message{
		ProductID: "123",
		Product: models.Product{
			ProductName: "Test Product",
		},
	}

	ctx := &gin.Context{}

	messageChan := make(chan models.Message, 1)

	// Start the goroutine for consuming messages
	go func() {
		err := mockProductService.ConsumeMessages(ctx, messageChan, mockReader)
		if err != nil {
			t.Errorf("Failed to consume messages: %v", err)
		}
	}()

	messageData, err := json.Marshal(expectedMessage)
	if err != nil {
		t.Errorf("Failed to marshal expected message: %v", err)
	}

	mockReader.ReceivedMessage = kafka.Message{
		Value: messageData,
	}

	// Wait for the message to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify the received message
	select {
	case receivedMessage := <-messageChan:
		if receivedMessage.ProductID != expectedMessage.ProductID {
			t.Errorf("Unexpected message received. Expected %s, got %s", expectedMessage.ProductID, receivedMessage.ProductID)
		}
	default:
		t.Error("No message received")
	}
}

func TestDownloadAndCompressProductImages(t *testing.T) {

	transactionID := uuid.New().String()
	ctx := &gin.Context{
		Request: &http.Request{
			Header: http.Header{
				constants.TransactionID: []string{transactionID},
			},
		},
	}

	utils.InitLogClient()

	productId := 123
	testMessage := models.Message{
		ProductID: fmt.Sprint(productId),
		Product: models.Product{
			ProductName: "Test Product",
		},
	}

	mockReader := &MockKafkaReader{}
	mockProductService := NewMockProductService(&db.MockPostgres{}, &MockKafkaWriter{}, mockReader)
	mockProductService.product = models.Product{
		ProductID:     &productId,
		ProductName:   "Test Product",
		ProductImages: []string{"url1", "url2"},
	}
	compressedImages, _ := mockProductService.DownloadAndCompressProductImages(ctx, testMessage)
	mockProductService.UpdateCompressedProductImages(ctx, testMessage.ProductID, compressedImages)
}

func TestResizeImage(t *testing.T) {
	// Create a mock Gin context
	ctx := &gin.Context{
		Request: &http.Request{
			Header: http.Header{
				constants.TransactionID: []string{"test-transaction-id"},
			},
		},
	}

	// Create a temporary directory for the test
	tmpDir, err := ioutil.TempDir("", "image-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create the input file path
	inputPath := filepath.Join(tmpDir, "input.jpg")

	// Create a test image with a size of 100x100 pixels
	testImage := image.NewRGBA(image.Rect(0, 0, 100, 100))
	outputFile, err := os.Create(inputPath)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}
	defer outputFile.Close()
	jpeg.Encode(outputFile, testImage, nil)

	// Create the expected output file path
	outputPath := filepath.Join(tmpDir, "output.jpg")

	// Create the product service instance
	service := &ProductService{}

	// Call the resizeImage method
	err = service.resizeImage(ctx, inputPath, outputPath, 50, 50)
	if err != nil {
		t.Fatalf("Failed to resize image: %v", err)
	}

	// Verify the output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output file does not exist")
	}

	// Verify the output file dimensions
	outputImage, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer outputImage.Close()

	img, _, err := image.Decode(outputImage)
	if err != nil {
		t.Fatalf("Failed to decode output image: %v", err)
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	if width != 50 || height != 50 {
		t.Errorf("Unexpected output image dimensions. Expected 50x50, got %dx%d", width, height)
	}
}

func TestDownloadAndCompressImage1(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{
			Header: http.Header{},
		},
	}

	tempImagePath := "/Users/ankitchahal/dev/go/src/github.com/ankit/project/message-quening-system/Images/DO-NOT-DELETE.jpg"

	responseCode := http.StatusOK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Equal(t, http.MethodGet, r.Method)

		// Send the response
		w.WriteHeader(responseCode)
	}))
	defer server.Close()

	url := server.URL
	productService := &ProductService{}
	err := productService.getImage(ctx, url, models.Message{ProductID: "24"}, 1, tempImagePath)
	assert.NoError(t, err)

	// Verify that the output file exists
	_, err = os.Stat(tempImagePath)
	assert.NoError(t, err)
}
