package middlewares

import (
	"backend/app/auth"
	"backend/app/models"
	"net/http"
)

func IsAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(auth.UserContextKey).(*models.User)
		if user != nil && user.Role == models.AdminRole {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}
	})
}
