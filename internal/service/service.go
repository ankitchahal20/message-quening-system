package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ankit/project/message-quening-system/cmd/producer"
	"github.com/ankit/project/message-quening-system/internal/db"
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
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

	//
	kafkaErr := publishProductID(*productDetails.ProductID)
	if kafkaErr != nil {
		//http.Error(w, "Failed to publish product_id", http.StatusInternalServerError)
		fmt.Println("kafkaErr : ", kafkaErr)
		//return
	}

	return productDetails, nil
}

func publishProductID(productID int) error {

	message := kafka.Message{
		Key:   []byte(fmt.Sprint(productID)),
		Value: nil,
	}

	err := producer.KafkaWriter.WriteMessages(context.Background(), message)
	if err != nil {
		return err
	}
	return nil
}
