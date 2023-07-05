package service

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ankit/project/message-quening-system/cmd/consumer"
	"github.com/ankit/project/message-quening-system/cmd/producer"
	"github.com/ankit/project/message-quening-system/internal/db"
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	resize "github.com/nfnt/resize"
	"github.com/segmentio/kafka-go"
)

var productClient *ProductService
var imageOutputDir string = "Images"

type ProductService struct {
	repo db.ProductService
	ctx  *gin.Context
}

func NewProductService(conn db.ProductService) *ProductService {
	productClient = &ProductService{
		repo: conn,
		ctx:  &gin.Context{},
	}
	return productClient
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

	// Send the product ID to the message channel
	message := models.Message{
		ProductID: fmt.Sprint(*productDetails.ProductID),
		Product:   productDetails,
	}
	utils.MessageChan <- message

	return productDetails, nil
}

func (service *ProductService) ProduceMessages(messageChan <-chan models.Message) {
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

func (service *ProductService) ConsumeMessages(messageChan chan<- models.Message) {
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

		// Download and compress the product images
		compressedImages := service.downloadAndCompressProductImages(service.ctx, receivedMessage)
		fmt.Println("compressedImages : ", compressedImages)
		// Update the database with the compressed_product_images
		productID, _ := strconv.Atoi(receivedMessage.ProductID)
		service.updateProductImages(service.ctx, productID, compressedImages)

		// Print the processed message
		log.Printf("Processed message: ProductID=%s, ProductName=%s\n", receivedMessage.ProductID, receivedMessage.Product.ProductName)
	}
}

func (service *ProductService) downloadAndCompressProductImages(ctx *gin.Context, msg models.Message) []string {
	// Simulate image compression process.
	productID, _ := strconv.Atoi(msg.ProductID)
	productImages, _ := service.getProductImages(ctx, productID)
	fmt.Printf("Product images compressed for product_id: %s %s\n", msg.ProductID, productImages)

	// Create the output directory if it doesn't exist
	err := os.MkdirAll(imageOutputDir, 0755)
	if err != nil {
		//return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Iterate over the product images and download/compress each image
	i := 1
	var imagesPath []string
	for _, imageURL := range productImages {
		imagePath, err := downloadAndCompressImage(imageURL, i)
		i++
		imagesPath = append(imagesPath, imagePath)
		if err != nil {
			//return //fmt.Errorf("failed to download and compress image: %w", err)
		}
	}

	return imagesPath
}

func (service *ProductService) getProductImages(ctx *gin.Context, productID int) ([]string, *producterror.ProductError) {

	images, err := service.repo.GetProductImages(ctx, productID)
	if err != nil {
		return []string{}, err
	}

	return images, nil
}

func downloadAndCompressImage(imageURL string, index int) (string, error) {
	// Create the output file path
	outputPath := filepath.Join(imageOutputDir, fmt.Sprintf("image%d.jpg", index))
	fmt.Println("outputPath ", outputPath)
	// Create the output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	fmt.Println("outputFile ", outputFile)
	defer outputFile.Close()

	// Download the image from the URL
	response, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer response.Body.Close()

	// Copy the image data to the output file
	_, err = io.Copy(outputFile, response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write image data: %w", err)
	}
	fmt.Println("outputFile ", outputFile)

	//Resize the image
	err = resizeImage(outputPath, outputPath, 50, 50)
	if err != nil {
		return "", fmt.Errorf("failed to resize image: %w", err)
	}

	fmt.Printf("Image downloaded and compressed: %s\n", outputPath)
	return outputPath, nil
}

func resizeImage(inputPath, outputPath string, width, height int) error {
	// Open the input file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer file.Close()

	// Decode the input image
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	// Calculate the target size while maintaining aspect ratio
	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()

	if width == 0 {
		width = imgWidth * height / imgHeight
	} else if height == 0 {
		height = imgHeight * width / imgWidth
	}

	// Resize the image using Lanczos resampling
	resizedImage := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	// Create the output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Encode the resized image and save it to the output file
	switch filepath.Ext(outputPath) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(outputFile, resizedImage, nil)
	case ".png":
		err = png.Encode(outputFile, resizedImage)
	default:
		err = fmt.Errorf("unsupported output format: %s", filepath.Ext(outputPath))
	}
	if err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}
	path, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Image resized and saved to %s\n", path+"/"+outputPath)
	return nil
}

func (service *ProductService) updateProductImages(ctx *gin.Context, productID int, compressedImages []string) {
	// Update the compressed_product_images column in the database
	fmt.Println("compressedImages : ", compressedImages)

	service.repo.UpdateCompressedProductImages(ctx, productID, compressedImages)
	fmt.Printf("Compressed images updated for product_id: %s\n", productID)
}
