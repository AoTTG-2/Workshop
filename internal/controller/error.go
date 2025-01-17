package controller

// APIError is echo.HTTPError wrapper for swagger generator
type APIError struct {
	Message string `json:"message"`
}
