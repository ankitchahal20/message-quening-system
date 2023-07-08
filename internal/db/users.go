package db

import (
	"net/http"
	"strings"

	"github.com/ankit/project/message-quening-system/internal/constants"
	"github.com/ankit/project/message-quening-system/internal/models"
	producterror "github.com/ankit/project/message-quening-system/internal/producterror"
	"github.com/ankit/project/message-quening-system/internal/utils"
	"github.com/gin-gonic/gin"
)

func (p postgres) AddUser(ctx *gin.Context, userDetails models.User) (*int, *producterror.ProductError) {
	query := `INSERT INTO users(name, mobile, latitude, longitude, created_at, updated_at) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`

	userID := 0
	err := p.db.QueryRow(query, userDetails.Name, userDetails.Mobile, userDetails.Latitude,
		userDetails.Longitude, userDetails.CreatedAt, userDetails.UpdatedAt).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return nil, &producterror.ProductError{
				Trace:   ctx.Request.Header.Get(constants.TransactionID),
				Code:    http.StatusBadRequest,
				Message: "user already added",
			}
		} else {
			return nil, &producterror.ProductError{
				Trace:   ctx.Request.Header.Get(constants.TransactionID),
				Code:    http.StatusInternalServerError,
				Message: "unable to add user details",
			}
		}
	}
	utils.Logger.Info("user added in db successfully")

	return &userID, nil
}
