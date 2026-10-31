package bind

import "github.com/labstack/echo/v4"

func PathParams[T any](c echo.Context) (T, error) {
	var value T
	err := new(echo.DefaultBinder).BindPathParams(c, &value)
	return value, err
}

func QueryParams[T any](c echo.Context) (T, error) {
	var value T
	err := new(echo.DefaultBinder).BindQueryParams(c, &value)
	return value, err
}

func Body[T any](c echo.Context) (T, error) {
	var value T
	err := new(echo.DefaultBinder).BindBody(c, &value)
	return value, err
}
