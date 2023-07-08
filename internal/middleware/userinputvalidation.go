package middleware

import (
	"net/http"

	"github.com/ankit/project/message-quening-system/internal/constants"
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

func ValidateUserInputRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// fetch the transactionID
		txid := getTransactionID(ctx)

		// validate the body params
		var userRequestFields models.User
		err := ctx.ShouldBindBodyWith(&userRequestFields, binding.JSON)
		if err != nil {
			utils.RespondWithError(ctx, http.StatusBadRequest, constants.InvalidBody)
			return
		}

		productError := validateUserInputRequest(txid, userRequestFields)
		if productError != nil {
			utils.RespondWithError(ctx, productError.Code, productError.Message)
			return
		}
		ctx.Next()
	}
}

// Note : We have just validate if the user fields are present or not. Data type validation is not here
func validateUserInputRequest(txid string, userRequestFields models.User) *producterror.ProductError {
	// if userRequestFields.ID == nil {
	// 	utils.Logger.Error("user id missing", zap.String("txid", txid))
	// 	return &producterror.ProductError{
	// 		Trace:   txid,
	// 		Code:    http.StatusBadRequest,
	// 		Message: "user id missing",
	// 	}
	// }

	if userRequestFields.Name == "" {
		utils.Logger.Error("user name missing", zap.String("txid", txid))
		return &producterror.ProductError{
			Trace:   txid,
			Code:    http.StatusBadRequest,
			Message: "user name missing",
		}
	}

	if userRequestFields.Mobile == "" {
		utils.Logger.Error("User mobile number is missing", zap.String("txid", txid))
		return &producterror.ProductError{
			Trace:   txid,
			Code:    http.StatusBadRequest,
			Message: "user mobile number is missing",
		}
	}

	if userRequestFields.Latitude == nil {
		utils.Logger.Error("latitude for user location is missing", zap.String("txid", txid))
		return &producterror.ProductError{
			Trace:   txid,
			Code:    http.StatusBadRequest,
			Message: "latitude for user location is missing",
		}
	}
	if userRequestFields.Longitude == nil {
		utils.Logger.Error("longitude for user location is missing", zap.String("txid", txid))
		return &producterror.ProductError{
			Trace:   txid,
			Code:    http.StatusBadRequest,
			Message: "longitude for user location is missing",
		}
	}

	return nil
}
