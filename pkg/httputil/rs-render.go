package httputil

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
)

// Renderer allows populating http response body with relevant content types.
// For now only JSON Response objects are supported.
type Renderer interface {
	// Render method renders data variables on the writer (most likely) http.ResponseWriter.
	Render(w io.Writer) error
	// Content-Type returns content type of the Renderer.
	ContentType() string
}

// JSONRend returns a concrete implementation of Renderer that can be used to populate JSON http response for given data.
func JSONRend(d interface{}) Renderer {
	return jsonRend{data: d}
}

// RsRender populates http response body where the behaviour is provideed by given renderer implmentation.
// Content-Type header will also be set as per renderer implementation.
func RsRender(w http.ResponseWriter, r Renderer) error {
	w.Header().Add("Content-Type", r.ContentType())
	if err := r.Render(w); err != nil {
		return status.ErrInternal().WithError(err)
	}
	return nil
}

// RsRenderWithStatus populates http response body where the behaviour is provideed by given renderer implmentation.
// Content-Type header will also be set as per renderer implementation.
// HTTP status code is set with given value.
func RsRenderWithStatus(w http.ResponseWriter, r Renderer, code int) error {
	w.Header().Add("Content-Type", r.ContentType())
	httpStatusCode := code
	w.WriteHeader(httpStatusCode)
	return RsRender(w, r)
}

type jsonRend struct {
	data interface{}
}

func (jr jsonRend) ContentType() string {
	return "application/json"
}
func (jr jsonRend) Render(w io.Writer) error {
	return json.NewEncoder(w).Encode(jr.data)
}
