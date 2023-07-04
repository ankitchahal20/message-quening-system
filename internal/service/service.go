package service

import (
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
		var urlInfo models.Product
		if err := context.ShouldBindBodyWith(&urlInfo, binding.JSON); err == nil {

			shortURL, err := productClient.createProduct(urlInfo)
			if err != nil {
				context.Writer.WriteHeader(http.StatusInternalServerError)
			} else {
				context.JSON(http.StatusCreated, shortURL)
				//context.Writer.WriteHeader(http.StatusOK)
			}
		} else {
			context.JSON(http.StatusBadRequest, gin.H{"Unable to marshal the request body": err.Error()})
		}
	}
}

func (service *ProductService) createProduct(urlInfo models.Product) (models.Product, *producterror.ProductError) {
	return models.Product{}, nil
}
