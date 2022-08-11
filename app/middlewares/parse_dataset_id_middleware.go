package middlewares

import (
	utils "backend/app/controllers/utils"
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

const DatasetIdContextKey ContextKey = "dataset_id"

func ParseDatasetIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		datasetIdString := vars["datasetId"]
		datasetId, err := strconv.Atoi(datasetIdString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			utils.WriteError(err, w)
			return
		}

		ctx := context.WithValue(r.Context(), DatasetIdContextKey, datasetId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
