package middlewares

import (
	"backend/app/auth"
	utils "backend/app/controllers/utils"
	"context"
	"errors"
	"net/http"
)

func AuthTokenMiddleware(tokenAuth *auth.TokenAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(auth.AuthTokenCookieName)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				utils.WriteError(errors.New("Unauthorized"), w)
				return
			}

			user, authErr := tokenAuth.CheckAuthToken(cookie.Value)
			if authErr != nil {
				w.WriteHeader(http.StatusUnauthorized)
				utils.WriteError(errors.New("Unauthorized"), w)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
