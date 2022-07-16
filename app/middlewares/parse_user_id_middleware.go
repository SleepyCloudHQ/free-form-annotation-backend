package middlewares

import (
	"context"
	"fmt"
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
			fmt.Println("Error converting user id")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := context.WithValue(r.Context(), UserIdContextKey, datasetId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
