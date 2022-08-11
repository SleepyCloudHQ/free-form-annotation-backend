package middlewares

import (
	utils "backend/app/controllers/utils"
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

const UserIdContextKey ContextKey = "user_id"

func ParseUserIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		datasetIdString := vars["userId"]
		datasetId, err := strconv.Atoi(datasetIdString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			utils.WriteError(err, w)
			return
		}

		ctx := context.WithValue(r.Context(), UserIdContextKey, datasetId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
