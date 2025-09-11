package handlers

import "github.com/gin-gonic/gin"

// SuccessResponse is used for successful responses
type SuccessResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

// ErrorResponse is used for error responses
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// JSONSuccess writes a successful JSON response
func JSONSuccess[T any](c *gin.Context, data T, status int) {
	c.JSON(status, SuccessResponse[T]{
		Success: true,
		Data:    data,
	})
}

// JSONError writes an error JSON response
func JSONError(c *gin.Context, message string, status int) {
	c.JSON(status, ErrorResponse{
		Success: false,
		Message: message,
	})
}
