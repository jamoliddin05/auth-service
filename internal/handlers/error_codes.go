package handlers

type ErrorCode string

const (
	ErrUserExists         ErrorCode = "ERR_USER_EXISTS"
	ErrInternal           ErrorCode = "ERR_INTERNAL"
	ErrInvalidCredentials ErrorCode = "ERR_INVALID_CREDENTIALS"
	ErrBadRequest         ErrorCode = "ERR_BAD_REQUEST"
	ErrRequired           ErrorCode = "ERR_REQUIRED"
	ErrMinLength          ErrorCode = "ERR_MIN_LENGTH"
	ErrMaxLength          ErrorCode = "ERR_MAX_LENGTH"
	ErrInvalidPhone       ErrorCode = "ERR_INVALID_PHONE"
	ErrInvalidField       ErrorCode = "ERR_INVALID_FIELD"
)
