package httputil

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type CtxKey int

const (
	CtxKeyRqID CtxKey = iota
	CtxKeyToken
	CtxKeyAuthSubj
)

func newCtxWithRqID(ctx context.Context, r *http.Request) context.Context {
	rqID := r.Header.Get("X-Request-ID")
	if rqID == "" {
		rqID = uuid.New().String()
	}
	return context.WithValue(ctx, CtxKeyRqID, rqID)
}

// CtxSubject returns subject that was stored against by AuthHandler for a HTTP Request.
// Subject is extracted from valid token claims.
func CtxSubject(ctx context.Context) string {
	v := ctx.Value(CtxKeyAuthSubj)
	if v == nil {
		return ""
	}
	return v.(string)
}

// CtxToken returns token that was stored against by AuthHandler for a HTTP Request.
func CtxToken(ctx context.Context) string {
	v := ctx.Value(CtxKeyToken)
	if v == nil {
		return ""
	}
	return v.(string)
}
