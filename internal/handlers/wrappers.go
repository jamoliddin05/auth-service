package handlers

import "github.com/gin-gonic/gin"

type APIResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Data    any               `json:"data,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

// ValidationResult represents the outcome of validation
type ValidationResult struct {
	Valid  bool
	Errors map[string]string
	Err    error
}

// JSONSuccess writes a success JSON response
func JSONSuccess(c *gin.Context, data any, statusCode int) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
	})
}

// JSONError writes an error JSON response. Message is optional.
func JSONError(c *gin.Context, statusCode int, errors map[string]string, message ...string) {
	var msg string
	if len(message) > 0 {
		msg = message[0]
	}

	resp := APIResponse{
		Success: false,
		Message: msg,
	}

	if len(errors) > 0 {
		resp.Errors = errors
	}

	c.JSON(statusCode, resp)
}
