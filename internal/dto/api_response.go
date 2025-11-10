package dto

type APIResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Data    Response          `json:"data,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}
