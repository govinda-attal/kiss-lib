package httputil

import (
	"net/http"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
)

// NotFoundHandler is a custom NOT Found handler for gorilla mux.
// It returns HTTP 404 Status along with custom JSON message - {msg: "Not Found: Resource path not mapped"}.
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	err := status.ErrNotFound().WithMessage("Resource path not mapped")
	RsRender(w, JSONRend(&err))
}
