package middlewares

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type ContextKey string

const DatasetIdContextKey ContextKey = "dataset_id"

func ParseDatasetIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		datasetIdString := vars["datasetId"]
		datasetId, err := strconv.Atoi(datasetIdString)
		if err != nil {
			fmt.Println("Error converting dataset id")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := context.WithValue(r.Context(), DatasetIdContextKey, datasetId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
