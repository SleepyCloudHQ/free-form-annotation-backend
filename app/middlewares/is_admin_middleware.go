package middlewares

import (
	"backend/app/auth"
	utils "backend/app/controllers/utils"
	"backend/app/models"
	"errors"
	"net/http"
)

func IsAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(auth.UserContextKey).(*models.User)
		if user != nil && user.Role == models.AdminRole {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			utils.WriteError(errors.New("Unauthorized"), w)
			return
		}
	})
}
