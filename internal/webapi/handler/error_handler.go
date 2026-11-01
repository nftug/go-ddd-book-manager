package handler

import (
	"errors"
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/webapi/auth"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
)

func ErrorHandler(c *echo.Context, err error) {
	attrs := []any{
		"method", c.Request().Method,
		"route", c.Path(),
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
	}

	response, responseErr := echo.UnwrapResponse(c.Response())
	if responseErr != nil {
		slog.ErrorContext(
			c.Request().Context(),
			"could not inspect response",
			append(attrs, "error", responseErr)...)
	}
	if response != nil && response.Committed {
		slog.ErrorContext(
			c.Request().Context(),
			"request failed after response was committed",
			append(attrs, "status", response.Status, "error", err)...,
		)
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

	attrs = append(attrs, "status", status)
	if status >= http.StatusInternalServerError {
		slog.ErrorContext(c.Request().Context(), "request failed", append(attrs, "error", err)...)
	} else {
		slog.WarnContext(c.Request().Context(), "request failed", attrs...)
	}

	if status == http.StatusUnauthorized {
		c.Response().Header().Set(echo.HeaderWWWAuthenticate, "Bearer")
	}

	if c.Request().Method == http.MethodHead {
		_ = c.NoContent(status)
		return
	}

	body := map[string]any{"message": message}
	if validation != nil {
		body["field"] = validation.Field()
		body["code"] = validation.Code()
	}

	_ = c.JSON(status, body)
}
