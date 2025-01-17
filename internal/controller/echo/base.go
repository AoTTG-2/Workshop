package echo

import (
	"github.com/labstack/echo/v4"
	"workshop/internal/controller"
)

type baseHandler struct {
}

func (h *baseHandler) getUser(ctx echo.Context) *controller.User {
	return ctx.Get("user").(*controller.User)
}

func (h *baseHandler) bindAndValidate(ctx echo.Context, req any) error {
	if err := ctx.Bind(req); err != nil {
		return err
	}

	if err := ctx.Validate(req); err != nil {
		return err
	}

	return nil
}
