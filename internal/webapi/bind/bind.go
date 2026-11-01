package bind

import "github.com/labstack/echo/v5"

func PathParams[T any](c *echo.Context) (T, error) {
	var value T
	err := echo.BindPathValues(c, &value)
	return value, err
}

func QueryParams[T any](c *echo.Context) (T, error) {
	var value T
	err := echo.BindQueryParams(c, &value)
	return value, err
}

func Body[T any](c *echo.Context) (T, error) {
	var value T
	err := echo.BindBody(c, &value)
	return value, err
}
