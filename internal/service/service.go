package service

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

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
	repo   db.ProductDBService
	writer KafkaWriter
	reader KafkaReader
}

type KafkaWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}
type KafkaReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
}

func NewProductService(conn db.ProductDBService, writer KafkaWriter, reader KafkaReader) *ProductService {
	productClient = &ProductService{
		repo:   conn,
		writer: writer,
		reader: reader,
	}
	return productClient
}

func AddProduct() func(ctx *gin.Context) {
	return func(context *gin.Context) {
		var productDetails models.Product
		if err := context.ShouldBindBodyWith(&productDetails, binding.JSON); err == nil {
			messageChan = make(chan models.Message)
			//var wg sync.WaitGroup
			//wg.Add(2)
			go func() {
				//defer wg.Done()
				err := productClient.ProduceMessages(context, messageChan, productClient.writer)
				if err != nil {
					utils.Logger.Error("Error producing messages:", zap.Error(err))
					context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to produce messages"})
				}
			}()

			go func() {
				//defer wg.Done()
				err := productClient.ConsumeMessages(context, messageChan, productClient.reader)
				if err != nil {
					utils.Logger.Error("Error consuming messages:", zap.Error(err))
					context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to consume messages"})
				}
			}()

			productID, err := productClient.createProduct(context, productDetails)
			//wg.Wait()
			if err != nil {
				context.JSON(err.Code, err)
			} else {
				context.JSON(http.StatusOK, map[string]string{
					//"Product Name": product.ProductName,
					"Product ID": fmt.Sprint(*productID),
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

func (service *ProductService) createProduct(ctx *gin.Context, productDetails models.Product) (*int, *producterror.ProductError) {
	productDetails.CreatedAt = time.Now().UTC()
	productDetails.UpdatedAt = time.Now().UTC()

	utils.Logger.Info("calling db layer for creating the product")
	productID, err := service.repo.AddProduct(ctx, productDetails)
	if err != nil {
		return nil, err
	}

	// Send the product ID to the message channel
	message := models.Message{
		ProductID: fmt.Sprint(*productID),
	}
	messageChan <- message

	return productID, nil
}

func (service *ProductService) ProduceMessages(ctx *gin.Context, messageChan <-chan models.Message, writer KafkaWriter) error {
	for message := range messageChan {
		// Serialize the message data
		messageData, err := json.Marshal(message)
		if err != nil {
			utils.Logger.Error("Error marshaling message :", zap.String("error", err.Error()))
			continue
		}
		utils.Logger.Info(fmt.Sprintf("Producer successfully marshalls the productId %v", message.ProductID))

		// Create a Kafka message with the serialized data
		kafkaMessage := kafka.Message{
			Key:   []byte(message.ProductID),
			Value: messageData,
		}

		// Write the message to the Kafka topic
		err = writer.WriteMessages(ctx, kafkaMessage)
		if err != nil {
			utils.Logger.Error("Error writing message to Kafka:", zap.String("error", err.Error()))
		}
		utils.Logger.Info(fmt.Sprintf("Producer successfully puts the productId %v on message queue", message.ProductID))
	}
	return nil
}

func (service *ProductService) ConsumeMessages(ctx *gin.Context, messageChan chan<- models.Message, reader KafkaReader) error {
	for {
		// Read the next message from the Kafka topic
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			utils.Logger.Error("Error reading message from Kafka:", zap.String("error", err.Error()))
			return fmt.Errorf("error reading message from Kafka: %w", err)
		}
		utils.Logger.Info("Consumser successfully reads the message from message queue")

		// Deserialize the kafka message data into a Message struct
		receivedMessage := models.Message{}
		err = json.Unmarshal(message.Value, &receivedMessage)
		if err != nil {
			utils.Logger.Error("Error marshaling message :", zap.String("error", err.Error()))
			return fmt.Errorf("error unmarshaling message: %w", err)
		}
		utils.Logger.Info(fmt.Sprintf("Consumser successfully unmarshalls the message, ProductId : %v", receivedMessage.ProductID))

		// Download and compress the product images
		compressedImages, productErr := service.downloadAndCompressProductImages(ctx, receivedMessage)
		if productErr != nil {
			utils.Logger.Error("unable to download and compress images :", zap.String("error", err.Error()))
			return fmt.Errorf("error downloading and compressing images: %v", productErr)
		}

		utils.Logger.Info(fmt.Sprintf("Consumser has successfully downloaded and compress the images for productId : %v", receivedMessage.ProductID))

		// Update the database with the compressed_product_images
		productID, _ := strconv.Atoi(receivedMessage.ProductID)
		producterr := service.updateCompressedProductImages(ctx, productID, compressedImages)
		if producterr != nil {
			utils.Logger.Error("unable to update compress images in db :", zap.String("error", err.Error()))
			return fmt.Errorf("error updating compressed images in db: %v", productErr)
		}
		utils.Logger.Info(fmt.Sprintf("Consumser has successfully updated the db with compressed images path for productId : %v", receivedMessage.ProductID))
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

	// Define the output file path

	// Iterate over the product images and download/compress each image
	i := 1
	var imagesPath []string
	var outputPath string
	for _, imageURL := range productImages {

		outputPath = filepath.Join(imageOutputDir, fmt.Sprintf(msg.ProductID+"-"+"image-%d.jpg", i))

		err := service.getImage(ctx, imageURL, msg, i, outputPath)
		i++

		if err != nil {
			utils.Logger.Error("failed to download and compress image", zap.String("error", err.Error()))
			return imagesPath, &producterror.ProductError{
				Code:    http.StatusInternalServerError,
				Message: "failed to download and compress image",
				Trace:   ctx.Request.Header.Get(constants.TransactionID),
			}
		}

		err = service.resizeImage(ctx, outputPath, outputPath, 50, 50)
		if err != nil {
			utils.Logger.Error("failed to resize image", zap.String("error", err.Error()))
			return imagesPath, &producterror.ProductError{
				Code:    http.StatusInternalServerError,
				Message: "failed to resize image",
				Trace:   ctx.Request.Header.Get(constants.TransactionID),
			}
		}

		path, err := os.Getwd()
		if err != nil {
			return imagesPath, &producterror.ProductError{
				Code:    http.StatusInternalServerError,
				Message: "failed to pwd path",
				Trace:   ctx.Request.Header.Get(constants.TransactionID),
			}
		}

		imagePath := path + "/" + outputPath

		imagesPath = append(imagesPath, imagePath)
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

func (service *ProductService) getImage(ctx *gin.Context, imageURL string, msg models.Message, index int, outputPath string) error {
	txid := ctx.Request.Header.Get(constants.TransactionID)

	// Create the output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		utils.Logger.Error("failed to create output file", zap.String("error", err.Error()), zap.String("txid", txid))
		return fmt.Errorf("failed to create output file: %w", err)
	}

	defer outputFile.Close()

	// Download the image from the URL
	response, err := http.Get(imageURL)
	if err != nil {
		utils.Logger.Error("failed to download image", zap.String("error", err.Error()), zap.String("txid", txid))
		return fmt.Errorf("failed to download image: %w", err)
	}
	defer response.Body.Close()

	// Copy the image data to the output file
	_, err = io.Copy(outputFile, response.Body)
	if err != nil {
		utils.Logger.Error("failed to write image data", zap.String("error", err.Error()), zap.String("txid", txid))
		return fmt.Errorf("failed to write image data: %w", err)
	}

	return nil

}

func (service *ProductService) resizeImage(ctx *gin.Context, inputPath, outputPath string, width, height int) error {
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
	utils.Logger.Info("calling db layer to update compressed product images")
	err := service.repo.UpdateCompressedProductImages(ctx, productID, compressedImages)
	return err
}
