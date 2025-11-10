package middlewares

import (
	"app/internal/dto"
	"app/internal/validators"
	"github.com/gin-gonic/gin"
	"net/http"
)

type RequestValidator struct {
	validator validators.Validator
}

func NewRequestValidator(validator validators.Validator) *RequestValidator {
	return &RequestValidator{validator: validator}
}

func (r *RequestValidator) ValidateRequest(c *gin.Context, req dto.Request) bool {
	var errResp dto.APIResponse
	errResp.Success = false
	errResp.Errors = make(map[string]string)

	err := c.BindJSON(req)
	if err != nil {
		errResp.Errors["error"] = err.Error()
		c.JSON(http.StatusBadRequest, errResp)
		return false
	}

	vr := r.validator.Validate(req)
	if !vr.Valid {
		errResp.Errors = vr.Errors
		c.JSON(http.StatusBadRequest, errResp)
		return false
	}

	return true
}
