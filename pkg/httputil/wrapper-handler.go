package httputil

import (
	"net/http"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
)

// HandlerFunc describes a signature for application specific handlers.
// This signature is different to standard http handler and intentionally simplifies error handling.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// DecoratorFunc can be used to add some decorations around real Handlers
type DecoratorFunc func(f HandlerFunc) HandlerFunc

// WrapperHandler is wrapper function to wrap API handlers and retuns as http.HandlerFunc.
// API Handlers may return error, and this wrapper simplifies error handling for API Handlers.
func WrapperHandler(f HandlerFunc, dd ...DecoratorFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := newCtxWithRqID(r.Context(), r)
		hf := f
		for _, d := range dd {
			hf = d(hf)
		}
		err := hf(w, r.WithContext(ctx))
		if err != nil {
			errProcessor(err, w)
		}
	}
}

func errProcessor(err error, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	if errSvc, ok := err.(status.ErrServiceStatus); ok {
		RsRenderWithStatus(w, JSONRend(&errSvc), errSvc.Code.HTTPStatusCode())
		return
	}
	errSvc := status.ErrInternal().WithMessage(err.Error())
	RsRender(w, JSONRend(&errSvc))
}
