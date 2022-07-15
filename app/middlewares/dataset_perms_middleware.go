package middlewares

import (
	"backend/app/auth"
	"backend/app/models"
	"gorm.io/gorm"
	"net/http"
)

func GetDatasetPermsMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := r.Context().Value(auth.UserContextKey).(*models.User)
			// admin users have perms to all the datasets
			if user.Role == models.AdminRole {
				next.ServeHTTP(w, r)
				return
			}

			datasetId := uint(r.Context().Value(DatasetIdContextKey).(int))

			// check if the user has been assigned to the given dataset
			userDataset := &models.UserDataset{
				UserID:    user.ID,
				DatasetID: datasetId,
			}

			if result := db.First(userDataset); result.Error != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
