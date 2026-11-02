package handler

import (
	"errors"
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/webapi/auth"

	"net/http"

	"github.com/labstack/echo/v5"
)

// ErrorResponse documents the JSON error response.
type ErrorResponse struct {
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
	Code    string `json:"code,omitempty"`
} // @name ErrorResponse

func ErrorHandler(c *echo.Context, err error) {
	response, _ := echo.UnwrapResponse(c.Response())
	if response != nil && response.Committed {
		return
	}

	status := http.StatusInternalServerError
	message := "Internal Server Error"
	var validation *dc.ValidationError

	switch {
	case errors.Is(err, auth.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = "Unauthorized"
	case errors.Is(err, dc.ErrForbidden):
		status = http.StatusForbidden
		message = "Forbidden"
	case errors.Is(err, dc.ErrNotFound):
		status = http.StatusNotFound
		message = "Not Found"
	case errors.Is(err, dc.ErrConflict):
		status = http.StatusConflict
		message = "Conflict"
	case errors.As(err, &validation):
		status = http.StatusBadRequest
		message = "Invalid input"
	case echo.StatusCode(err) != 0:
		status = echo.StatusCode(err)
		if status < http.StatusInternalServerError {
			message = http.StatusText(status)
		}
	}

	if status == http.StatusUnauthorized {
		c.Response().Header().Set(echo.HeaderWWWAuthenticate, "Bearer")
	}

	if c.Request().Method == http.MethodHead {
		_ = c.NoContent(status)
		return
	}

	body := ErrorResponse{Message: message}
	if validation != nil {
		body.Field = validation.Field()
		body.Code = validation.Code()
	}

	_ = c.JSON(status, body)
}
