package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
	userID := 1
	mockReader := &MockKafkaReader{}
	mockProductService := NewMockProductService(&db.MockPostgres{}, &MockKafkaWriter{}, mockReader)
	mockProductService.product = models.Product{
		UserID:        &userID,
		ProductName:   "Test Product",
		ProductImages: []string{"url1", "url2"},
	}
	compressedImages, _ := mockProductService.DownloadAndCompressProductImages(ctx, testMessage)
	mockProductService.UpdateCompressedProductImages(ctx, testMessage.ProductID, compressedImages)
}

func TestDownloadAndCompressImage(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{
			Header: http.Header{},
		},
	}

	tempImagePath := "/Users/ankitchahal/dev/go/src/github.com/ankit/project/message-quening-system/cmd/Images/DO-NOT-DELETE.jpg"

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
