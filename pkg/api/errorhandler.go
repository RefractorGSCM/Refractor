package api

import (
	"Refractor/domain"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
)

func GetEchoErrorHandler(logger *zap.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		// If this error is a custom http error, treat it as such
		if httpError, ok := err.(*domain.HTTPError); ok {

			body, err := httpError.ResponseBody()
			if err != nil {
				c.Logger().Error(err)
				return
			}

			code, _ := httpError.ResponseHeaders()

			err = c.JSONBlob(code, body)
			if err != nil {
				c.Logger().Error(err)
				return
			}
		} else if echoErr, ok := err.(*echo.HTTPError); ok {
			err := c.JSON(echoErr.Code, domain.Response{
				Success: false,
				Message: echoErr.Message.(string),
			})
			if err != nil {
				c.Logger().Error(err)
				return
			}
			return
		} else {
			// If this error is not a custom http error, assume it's an internal error.
			// We log it and then send back an internal server error message to the user.
			logger.Error("An error occurred", zap.Error(err))

			type intErr struct {
				Message string `json:"message"`
			}

			err := c.JSON(http.StatusInternalServerError, intErr{Message: "Internal server error"})
			if err != nil {
				c.Logger().Error(err)
				return
			}
		}
	}
}
