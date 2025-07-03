package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"net/http"
	"salesforge-assignment/internal/api"
	apierrors "salesforge-assignment/internal/api-errors"
	"salesforge-assignment/internal/logger"
)

func HandleError(c *gin.Context, err error) {
	log := logger.FromContext(c)

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		log.Error().Err(err).Msg("Validation error occurred")
		handleValidationError(c, ve)
		return
	} else {
		log.Error().Err(err).Msg("Error occurred while handling request")
		handleDefaultError(c, err)
	}

}

func handleValidationError(c *gin.Context, ve validator.ValidationErrors) {
	details := api.ValidationErrors{}

	for _, fe := range ve {
		errorDetail := api.ValidationError{
			Field:   fe.Field(),
			Message: fmt.Sprint("Field validation for '", fe.Tag(), "' failed."),
		}

		details = append(details, errorDetail)
	}

	validationResponse := api.ValidationErrorResponse{
		Code:    400,
		Message: "One or more fields failed validation.",
		Errors:  details,
	}
	c.JSON(http.StatusBadRequest, validationResponse)
	return
}

func handleDefaultError(c *gin.Context, err error) {
	var httpErr apierrors.HTTPError
	if errors.As(err, &httpErr) {
		apiErr := httpErr.APIErrorResponse()
		c.JSON(apiErr.Code, apiErr)
		return
	} else {
		log.Error().Err(err).Msg("Unhandled internal server error")
		apiErr := apierrors.InvalidApplicationStateError{}
		response := apiErr.APIErrorResponse()
		c.JSON(response.Code, response)
	}

}
