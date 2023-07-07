package middleware

import (
	"net/http"

	"github.com/ankit/project/message-quening-system/internal/constants"
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// This function gets the unique transactionID
func getTransactionID(c *gin.Context) string {

	transactionID := c.GetHeader(constants.TransactionID)
	_, err := uuid.Parse(transactionID)
	if err != nil {
		transactionID = uuid.New().String()
		c.Request.Header.Set(constants.TransactionID, transactionID)
	}
	return transactionID
}

func ValidateInputRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// fetch the transactionID
		txid := getTransactionID(ctx)

		// validate the body params
		var productRequestFields models.Product
		err := ctx.ShouldBindBodyWith(&productRequestFields, binding.JSON)
		if err != nil {
			utils.RespondWithError(ctx, http.StatusBadRequest, constants.InvalidBody)
			return
		}

		productError := validateInputRequest(txid, productRequestFields)
		if productError != nil {
			utils.RespondWithError(ctx, productError.Code, productError.Message)
			return
		}
		ctx.Next()
	}
}

func validateInputRequest(txid string, productRequestFields models.Product) *producterror.ProductError {
	if productRequestFields.UserID == nil {
		utils.Logger.Error("user id missing", zap.String("txid", txid))
		return &producterror.ProductError{
			Trace:   txid,
			Code:    http.StatusBadRequest,
			Message: "user id missing",
		}
	}

	if productRequestFields.ProductName == "" {
		utils.Logger.Error("Product name missing", zap.String("txid", txid))
		return &producterror.ProductError{
			Trace:   txid,
			Code:    http.StatusBadRequest,
			Message: "Product name missing",
		}
	}

	if productRequestFields.ProductDescription == "" {
		utils.Logger.Error("Product description  missing", zap.String("txid", txid))
		return &producterror.ProductError{
			Trace:   txid,
			Code:    http.StatusBadRequest,
			Message: "Product description missing",
		}
	}

	if productRequestFields.ProductPrice == nil {
		utils.Logger.Error("Product price missing", zap.String("txid", txid))
		return &producterror.ProductError{
			Trace:   txid,
			Code:    http.StatusBadRequest,
			Message: "Product price missing",
		}
	}
	if len(productRequestFields.ProductImages) == 0 {
		utils.Logger.Error("Product images missing", zap.String("txid", txid))
		return &producterror.ProductError{
			Trace:   txid,
			Code:    http.StatusBadRequest,
			Message: "Product images missing",
		}
	}

	return nil
}
