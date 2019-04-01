package codes

import (
	"net/http"
)

type Code int

const (
	// Success represents a generic success.
	Success Code = iota
	// ErrInternal represents internal server error.
	ErrInternal
	// ErrUnauthorized represents an unauthorized request error.
	ErrUnauthorized
	// ErrBadRequest represents an invalid request error.
	ErrBadRequest
	// ErrNotImplemented represents an unauthorized request error.
	ErrNotImplemented
	// ErrNotFound represents an error when a domain artifact was not found.
	ErrNotFound
	// ErrContentTypeNotSupported represents unsupported media type.
	ErrContentTypeNotSupported
	// ErrStatusConflict represents conflict because of inconsistent or duplicated info.
	ErrStatusConflict
)

func (c Code) HTTPStatusCode() int {
	switch c {
	case Success:
		return http.StatusOK
	case ErrInternal:
		return http.StatusInternalServerError
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrBadRequest:
		return http.StatusBadRequest
	case ErrNotImplemented:
		return http.StatusNotImplemented
	case ErrNotFound:
		return http.StatusNotFound
	case ErrContentTypeNotSupported:
		return http.StatusUnsupportedMediaType
	case ErrStatusConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
