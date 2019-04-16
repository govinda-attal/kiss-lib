package status

import (
	"encoding/json"
	"strings"

	"github.com/jinzhu/copier"

	"github.com/govinda-attal/kiss-lib/pkg/core/status/codes"
)

// ErrInternal represents internal server error.
func ErrInternal() ErrServiceStatus {
	return ErrServiceStatus{
		ServiceStatus{Code: codes.ErrInternal, Message: "Internal Server Error"}, nil,
	}
}

// ErrUnauthorized represents an unauthorized request error.
func ErrUnauthorized() ErrServiceStatus {
	return ErrServiceStatus{
		ServiceStatus{Code: codes.ErrUnauthorized, Message: "Unauthorized"}, nil,
	}
}

// ErrNotFound represents an error when a domain artifact was not found.
func ErrNotFound() ErrServiceStatus {
	return ErrServiceStatus{
		ServiceStatus{Code: codes.ErrNotFound, Message: "Not Found"}, nil,
	}
}

// ErrBadRequest represents an invalid request error.
func ErrBadRequest() ErrServiceStatus {
	return ErrServiceStatus{
		ServiceStatus{Code: codes.ErrBadRequest, Message: "Bad Request"}, nil,
	}
}

// ErrNotImplemented represents an unauthorized request error.
func ErrNotImplemented() ErrServiceStatus {
	return ErrServiceStatus{
		ServiceStatus{Code: codes.ErrNotImplemented, Message: "Not Implemented"}, nil,
	}
}

// ErrContentTypeNotSupported represents unsupported media type.
func ErrContentTypeNotSupported() ErrServiceStatus {
	return ErrServiceStatus{
		ServiceStatus{Code: codes.ErrContentTypeNotSupported, Message: "Unsupported Media Type"}, nil,
	}
}

// ErrStatusConflict represents conflict because of inconsistent or duplicated info.
func ErrStatusConflict() ErrServiceStatus {
	return ErrServiceStatus{
		ServiceStatus{Code: codes.ErrStatusConflict, Message: "Conflict because of inconsistent or duplicated info"}, nil,
	}
}

func ErrCause(err error) error {
	if e, ok := err.(errCauser); ok {
		return e.Cause()
	}
	return nil
}

// Success represents a generic success.
func Success() ServiceStatus {
	return ServiceStatus{Code: codes.Success, Message: "OK"}
}

// ServiceStatus captures basic information about a status construct.
type ServiceStatus struct {
	Code    codes.Code   `json:"code,omitempty"`
	Message string       `json:"msg,omitempty"`
	Details []*StatusDtl `json:"details,omitempty"`
}

// StatusDtl captures basic information about a status construct.
type StatusDtl struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"msg,omitempty"`
}

// ErrServiceStatus captures basic information about an error.
type ErrServiceStatus struct {
	ServiceStatus
	err error
}

// WithMessage returns an error status with given message.
func (e ErrServiceStatus) WithMessage(msg string) ErrServiceStatus {
	var err ErrServiceStatus
	copier.Copy(&err, e)
	err.Message = strings.Join([]string{e.Message, msg}, ": ")
	return err
}

// WithError returns an error status with given err.Error().
func (e ErrServiceStatus) WithError(err error) ErrServiceStatus {
	ex := e.WithMessage(err.Error())
	ex.err = err
	return ex
}

// AddDtlMsg returns an error status with given message.
func (e *ErrServiceStatus) AddDtlMsg(msgs ...string) {
	for _, m := range msgs {
		d := &StatusDtl{Message: m}
		e.Details = append(e.Details, d)
	}
}

// AddDtl returns an error status with given message.
func (e *ErrServiceStatus) AddDtl(code, msg string) {
	d := &StatusDtl{Code: code, Message: msg}
	e.Details = append(e.Details, d)
}

// NewUserDefined returns a new status with given code and message.
func NewUserDefined(code codes.Code, msg string) ServiceStatus {
	return ServiceStatus{Code: code, Message: msg}
}

func (e ErrServiceStatus) Error() string {
	if errB, err := json.Marshal(&e); err == nil {
		return string(errB)
	}
	return `{"code":500, "msg": "error marshal failed"}`
}

// Cause returns an error causer.
func (e ErrServiceStatus) Cause() error {
	return e.err
}

type errCauser interface {
	Cause() error
}

func (e ErrServiceStatus) Is(code codes.Code) bool {
	return e.Code == code
}
