package handler

import (
	"errors"
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/webapi/auth"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

func ErrorHandler(err error, c echo.Context) {
	attrs := []any{
		"method", c.Request().Method,
		"route", c.Path(),
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
	}

	if c.Response().Committed {
		slog.ErrorContext(
			c.Request().Context(),
			"request failed after response was committed",
			append(attrs, "status", c.Response().Status, "error", err)...,
		)
		return
	}

	status := 500
	message := "Internal Server Error"
	var validation *dc.ValidationError
	var httpErr *echo.HTTPError

	switch {
	case errors.Is(err, auth.ErrUnauthorized):
		status = 401
		message = "Unauthorized"
	case errors.Is(err, dc.ErrForbidden):
		status = 403
		message = "Forbidden"
	case errors.Is(err, dc.ErrNotFound):
		status = 404
		message = "Not Found"
	case errors.Is(err, dc.ErrConflict):
		status = 409
		message = "Conflict"
	case errors.As(err, &validation):
		status = 400
		message = "Invalid input"
	case errors.As(err, &httpErr):
		status = httpErr.Code
		message = http.StatusText(status)
	}

	attrs = append(attrs, "status", status)
	if status >= http.StatusInternalServerError {
		slog.ErrorContext(c.Request().Context(), "request failed", append(attrs, "error", err)...)
	} else {
		slog.WarnContext(c.Request().Context(), "request failed", attrs...)
	}

	if status == 401 {
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
