package router

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type EchoRouter struct {
	*echo.Echo
}

func NewEchoRouter(debug bool) *EchoRouter {
	r := new(EchoRouter)
	r.Echo = echo.New()
	r.Debug = debug
	return r
}

func (r *EchoRouter) WithDefaultNotFoundHandler() *EchoRouter {
	r.RouteNotFound("/*", func(c echo.Context) error {
		return c.NoContent(http.StatusNotFound)
	})
	return r
}

func (r *EchoRouter) WithValidator(validator echo.Validator) *EchoRouter {
	r.Validator = validator
	return r
}

func (r *EchoRouter) Run(port uint16) error {
	return r.Start(fmt.Sprintf(":%d", port))
}

func (r *EchoRouter) Stop(ctx context.Context) error {
	return r.Shutdown(ctx)
}
