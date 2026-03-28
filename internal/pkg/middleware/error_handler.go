package middleware

import (
	"errors"
	"net/http"

	"github.com/degeta10/workout-assistant-api/internal/pkg/apperrors"
	"github.com/degeta10/workout-assistant-api/internal/pkg/responses"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorHandler centralizes API error responses for handlers that use c.Error(err).
func ErrorHandler() gin.HandlerFunc {
	internalErrorMessage := "Something went wrong"
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Written() || len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		if err == nil {
			responses.InternalError(c, internalErrorMessage)
			return
		}

		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			responses.ValidationError(c, err)
			return
		}

		if appErr, ok := apperrors.As(err); ok {
			if appErr.StatusCode == http.StatusUnprocessableEntity {
				responses.ValidationError(c, appErr.Err)
				return
			}

			if appErr.Message == "" {
				responses.InternalError(c, internalErrorMessage)
				return
			}

			c.JSON(appErr.StatusCode, responses.APIResponse{
				Success: false,
				Message: appErr.Message,
				Errors: gin.H{
					"code":       appErr.Code,
					"request_id": GetRequestID(c),
				},
			})
			return
		}

		responses.InternalError(c, internalErrorMessage)
	}
}
