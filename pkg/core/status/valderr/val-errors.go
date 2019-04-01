package valderr

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/jinzhu/copier"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
)

func NewErrStatusWithValErrors(e status.ErrServiceStatus, valErrs validation.Errors) status.ErrServiceStatus {
	var errSvc status.ErrServiceStatus
	copier.Copy(&errSvc, e)
	for _, msg := range valErrs {
		errSvc.AddDtlMsg(msg.Error())
	}
	return errSvc
}
