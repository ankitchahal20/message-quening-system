package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ankit/project/message-quening-system/cmd/producer"
	"github.com/segmentio/kafka-go"
)

var KafkaReader *kafka.Reader

func IntializeKafkaConsumerReader() {
	KafkaReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{producer.Broker1Address},
		Topic:    producer.Topic,
		GroupID:  "my-group",
		MaxBytes: 1e6,
		//// if you set it to `kafka.LastOffset` it will only consume new messages
		StartOffset: kafka.LastOffset,
	})
}

func main() {
	IntializeKafkaConsumerReader()
	for {
		m, err := KafkaReader.ReadMessage(context.Background())
		if err != nil {
			break
		}
		productID := string(m.Key)

		// Process the product images and store them in the database.
		compressedImages := processProductImages(productID)

		// Update the database with the compressed_product_images.
		updateProductImages(productID, compressedImages)
	}

}

func processProductImages(productID string) []string {
	// Simulate image compression process.
	time.Sleep(time.Second)
	fmt.Printf("Product images compressed for product_id: %s\n", productID)

	// Return the compressed image file paths.
	return []string{
		"/path/to/compressed/image1.jpg",
		"/path/to/compressed/image2.jpg",
	}
}

func updateProductImages(productID string, compressedImages []string) {
	// Update the compressed_product_images column in the database.
	fmt.Printf("Compressed images updated for product_id: %s\n", productID)
}
