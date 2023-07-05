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
	"github.com/ankit/project/message-quening-system/internal/constants"
	"github.com/ankit/project/message-quening-system/internal/db"
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	resize "github.com/nfnt/resize"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var productClient *ProductService
var imageOutputDir string = "Images"
var messageChan chan models.Message

type ProductService struct {
	repo db.ProductService
}

func NewProductService(conn db.ProductService) *ProductService {
	productClient = &ProductService{
		repo: conn,
	}
	return productClient
}

func CreateProduct() func(ctx *gin.Context) {
	return func(context *gin.Context) {
		var productDetails models.Product
		if err := context.ShouldBindBodyWith(&productDetails, binding.JSON); err == nil {

			go productClient.ProduceMessages(context, messageChan)
			go productClient.ConsumeMessages(context, messageChan)

			product, err := productClient.createProduct(context, productDetails)
			if err != nil {
				context.Writer.Header().Set("Custom-Header", "Hello, World!")
				context.JSON(err.Code, err)
			} else {

				context.Writer.Header().Set("Custom-Header", "Hello, World!")
				context.JSON(http.StatusOK, map[string]string{
					"Product Name": product.ProductName,
					"Product ID":   fmt.Sprint(*product.ProductID),
				})
			}
		} else {
			producterror := producterror.ProductError{
				Code:    http.StatusBadRequest,
				Message: "unable to marshall the request body",
				Trace:   context.GetHeader(constants.TransactionID),
			}
			context.JSON(http.StatusBadRequest, producterror)
		}
	}
}

func (service *ProductService) createProduct(ctx *gin.Context, productDetails models.Product) (models.Product, *producterror.ProductError) {
	productDetails.CreatedAt = time.Now().UTC()
	productDetails.UpdatedAt = time.Now().UTC()

	utils.Logger.Info("calling db layer for creating the product")
	err := service.repo.CreateProduct(ctx, productDetails)
	if err != nil {
		return models.Product{}, err
	}

	// Send the product ID to the message channel
	message := models.Message{
		ProductID: fmt.Sprint(*productDetails.ProductID),
		Product:   productDetails,
	}
	messageChan <- message

	return productDetails, nil
}

func (service *ProductService) ProduceMessages(ctx *gin.Context, messageChan <-chan models.Message) {
	fmt.Println("Pro")
	for message := range messageChan {
		// Serialize the message data
		messageData, err := json.Marshal(message)
		if err != nil {
			utils.Logger.Error("Error marshaling message :", zap.String("error", err.Error()))
			continue
		}

		// Create a Kafka message with the serialized data
		kafkaMessage := kafka.Message{
			Key:   []byte(message.ProductID),
			Value: messageData,
		}

		// Write the message to the Kafka topic
		err = producer.KafkaWriter.WriteMessages(ctx, kafkaMessage)
		if err != nil {
			utils.Logger.Error("Error writing message to Kafka:", zap.String("error", err.Error()))
		}
	}
}

func (service *ProductService) ConsumeMessages(ctx *gin.Context, messageChan chan<- models.Message) {
	fmt.Println("Con")
	for {
		// Read the next message from the Kafka topic
		message, err := consumer.KafkaReader.ReadMessage(context.Background())
		if err != nil {
			utils.Logger.Error("Error reading message from Kafka:", zap.String("error", err.Error()))
			continue
		}

		// Deserialize the kafka message data into a Message struct
		var receivedMessage models.Message
		err = json.Unmarshal(message.Value, &receivedMessage)
		if err != nil {
			utils.Logger.Error("Error marshaling message :", zap.String("error", err.Error()))
			continue
		}

		// Download and compress the product images
		compressedImages, productErr := service.downloadAndCompressProductImages(ctx, receivedMessage)
		if productErr != nil {
			utils.Logger.Error("unable to download and compress images :", zap.String("error", err.Error()))
		}
		// Update the database with the compressed_product_images
		productID, _ := strconv.Atoi(receivedMessage.ProductID)
		producterr := service.updateCompressedProductImages(ctx, productID, compressedImages)
		if producterr != nil {
			utils.Logger.Error("unable to update compress images in db :", zap.String("error", err.Error()))
		}

		// Print the processed message
		log.Printf("Processed message: ProductID=%s, ProductName=%s\n", receivedMessage.ProductID, receivedMessage.Product.ProductName)
	}
}

func (service *ProductService) downloadAndCompressProductImages(ctx *gin.Context, msg models.Message) ([]string, *producterror.ProductError) {
	// Simulate image compression process.
	productID, _ := strconv.Atoi(msg.ProductID)
	productImages, _ := service.getProductImages(ctx, productID)
	fmt.Printf("Product images compressed for product_id: %s %s\n", msg.ProductID, productImages)

	// Create the output directory if it doesn't exist
	err := os.MkdirAll(imageOutputDir, 0755)
	if err != nil {
		utils.Logger.Error("failed to create output directory", zap.String("error", err.Error()))
		return []string{}, &producterror.ProductError{
			Code:    http.StatusInternalServerError,
			Message: "failed to create output directory",
			Trace:   ctx.Request.Header.Get(constants.TransactionID),
		}
	}

	// Iterate over the product images and download/compress each image
	i := 1
	var imagesPath []string
	for _, imageURL := range productImages {
		imagePath, err := downloadAndCompressImage(ctx, imageURL, i)
		i++
		imagesPath = append(imagesPath, imagePath)
		if err != nil {
			utils.Logger.Error("failed to download and compress image", zap.String("error", err.Error()))
			return imagesPath, &producterror.ProductError{
				Code:    http.StatusInternalServerError,
				Message: "failed to download and compress image",
				Trace:   ctx.Request.Header.Get(constants.TransactionID),
			}
		}
	}

	return imagesPath, nil
}

func (service *ProductService) getProductImages(ctx *gin.Context, productID int) ([]string, *producterror.ProductError) {

	images, err := service.repo.GetProductImages(ctx, productID)
	if err != nil {
		return []string{}, err
	}

	return images, nil
}

func downloadAndCompressImage(ctx *gin.Context, imageURL string, index int) (string, error) {
	txid := ctx.Request.Header.Get(constants.TransactionID)
	// Create the output file path
	outputPath := filepath.Join(imageOutputDir, fmt.Sprintf("image%d.jpg", index))

	// Create the output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		utils.Logger.Error("failed to create output file", zap.String("error", err.Error()), zap.String("txid", txid))
		return "", fmt.Errorf("failed to create output file: %w", err)
	}

	defer outputFile.Close()

	// Download the image from the URL
	response, err := http.Get(imageURL)
	if err != nil {
		utils.Logger.Error("failed to download image", zap.String("error", err.Error()), zap.String("txid", txid))
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer response.Body.Close()

	// Copy the image data to the output file
	_, err = io.Copy(outputFile, response.Body)
	if err != nil {
		utils.Logger.Error("failed to write image data", zap.String("error", err.Error()), zap.String("txid", txid))
		return "", fmt.Errorf("failed to write image data: %w", err)
	}

	//Resize the image
	err = resizeImage(ctx, outputPath, outputPath, 50, 50)
	if err != nil {
		utils.Logger.Error("failed to resize image", zap.String("error", err.Error()), zap.String("txid", txid))
		return "", fmt.Errorf("failed to resize image: %w", err)
	}

	path, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	utils.Logger.Info("Image downloaded and compressed: %s\n", zap.String("image path ", path+"/"+outputPath), zap.String("txid", txid))
	return path + "/" + outputPath, nil
}

func resizeImage(ctx *gin.Context, inputPath, outputPath string, width, height int) error {
	txid := ctx.Request.Header.Get(constants.TransactionID)
	// Open the input file
	file, err := os.Open(inputPath)
	if err != nil {
		utils.Logger.Error("failed to open input file", zap.String("error", err.Error()), zap.String("txid", txid))
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer file.Close()

	// Decode the input image
	img, _, err := image.Decode(file)
	if err != nil {
		utils.Logger.Error("failed to open input file", zap.String("error", err.Error()), zap.String("txid", txid))
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
		utils.Logger.Error("failed to create output file", zap.String("error", err.Error()), zap.String("txid", txid))
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
		utils.Logger.Error("unsupported output format", zap.String("error", err.Error()), zap.String("txid", txid))
		err = fmt.Errorf("unsupported output format: %s", filepath.Ext(outputPath))
	}
	if err != nil {
		utils.Logger.Error("failed to encode image", zap.String("error", err.Error()), zap.String("txid", txid))
		return fmt.Errorf("failed to encode image: %v", err)
	}

	utils.Logger.Info(fmt.Sprintf("Image resized and saved to %s\n", outputPath), zap.String("txid", txid))
	return nil
}

func (service *ProductService) updateCompressedProductImages(ctx *gin.Context, productID int, compressedImages []string) *producterror.ProductError {
	// Update the compressed_product_images column in the database
	utils.Logger.Error("calling db layer to update compressed product images")
	err := service.repo.UpdateCompressedProductImages(ctx, productID, compressedImages)
	return err
}
