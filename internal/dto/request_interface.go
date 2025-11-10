package dto

type Request interface {
	FieldErrorCode(field string) string
}
