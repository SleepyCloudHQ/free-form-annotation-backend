package middlewares

import (
	utils "backend/app/controllers/utils"
	"backend/app/handlers"
	"backend/app/models"
	"errors"
	"net/http"
)

func GetDatasetPermsMiddleware(handler *handlers.UserDatasetPermsHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContextKey).(*models.User)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// admin users have perms to all the datasets
			if user.Role == models.AdminRole {
				next.ServeHTTP(w, r)
				return
			}

			datasetId, datasetIdOk := r.Context().Value(DatasetIdContextKey).(int)
			if !datasetIdOk {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			hasPerm, permErr := handler.UserHasDatasetPerms(user.ID, uint(datasetId))
			if permErr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !hasPerm {
				w.WriteHeader(http.StatusUnauthorized)
				utils.WriteError(errors.New("Unauthorized"), w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
