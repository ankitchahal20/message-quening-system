package utils

import (
	"net/http"

	"github.com/ankit/project/message-quening-system/internal/constants"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLogClient() {
	Logger, _ = zap.NewDevelopment()
}

func RespondWithError(c *gin.Context, statusCode int, message string) {
	status := http.StatusText(statusCode)
	txnID := c.GetString(constants.TransactionID)
	c.AbortWithStatusJSON(statusCode, producterror.ProductError{
		Status:  &status,
		Trace:   txnID,
		Code:    statusCode,
		Message: message,
	})
}
