package httputil

import (
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
	"github.com/govinda-attal/kiss-lib/pkg/jwtkit"
)

// AuthDecorator can be applied to a specific path & HTTP verb combination.
// Ideally this is to be used when say one or few HTTP verbs require authentication and others don't on the same resource path.
func AuthDecorator(v jwtkit.Verifier) DecoratorFunc {
	return func(f HandlerFunc) HandlerFunc {
		return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			ctx := r.Context()
			authHdr := r.Header.Get("Authorization")
			if authHdr != "" && strings.Contains(authHdr, "Bearer ") {
				token := strings.Split(authHdr, "Bearer ")[1]
				ctx = context.WithValue(ctx, CtxKeyToken, token)
				claims, err := v.VerifyToken(token)
				if err != nil {
					errSvc := err.(status.ErrServiceStatus)
					return status.ErrUnauthorized().WithMessage(errSvc.Message)
				}
				sub := claims.(jwt.MapClaims)["sub"]
				ctx = context.WithValue(ctx, CtxKeyAuthSubj, sub)
				return f(w, r.WithContext(ctx))
			}
			return status.ErrUnauthorized().WithMessage("Authorization Bearer token is missing")
		})
	}
}
