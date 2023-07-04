package service

import (
	"fmt"
	"net/http"

	"github.com/ankit/project/message-quening-system/internal/db"
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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

			product, err := productClient.createProduct(productDetails)
			if err != nil {
				context.Writer.WriteHeader(http.StatusInternalServerError)
			} else {
				context.JSON(http.StatusCreated, product)
			}
		} else {
			context.JSON(http.StatusBadRequest, gin.H{"Unable to marshal the request body": err.Error()})
		}
	}
}

func (service *ProductService) createProduct(productDetails models.Product) (models.Product, *producterror.ProductError) {
	fmt.Println("hello from product service")
	return models.Product{}, nil
}
