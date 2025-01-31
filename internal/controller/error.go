package controller

import (
	"github.com/labstack/echo/v4"
	"time"
)

// APIError is echo.HTTPError wrapper for swagger generator
type APIError echo.HTTPError

type APIGenericError[T any] struct {
	Message string `json:"message" extensions:"x-order=0"`
	Data    T      `json:"data,omitempty" extensions:"x-order=1"`
}

type APIRateLimitErrorData struct {
	ResetAt time.Time `json:"reset_at"`
}
