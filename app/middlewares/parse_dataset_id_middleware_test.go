package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/matryer/is"
)

func TestParsingValidDatasetId(t *testing.T) {
	is := is.New(t)

	datasetId := 1

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextDatasetId, ok := r.Context().Value(DatasetIdContextKey).(int)
		is.True(ok)
		is.Equal(datasetId, contextDatasetId)
		w.WriteHeader(http.StatusOK)
	})

	testHandler := ParseDatasetIdMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)

	vars := map[string]string{
		DatasetIdVarKey: strconv.Itoa(datasetId),
	}

	testHandler.ServeHTTP(httptest.NewRecorder(), mux.SetURLVars(req, vars))
}

func TestParsingInvalidDatasetId(t *testing.T) {
	is := is.New(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should be stopped by the middleware anod it should not reach here")
	})

	testHandler := ParseDatasetIdMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)

	vars := map[string]string{
		DatasetIdVarKey: "this is not an id",
	}

	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, mux.SetURLVars(req, vars))
	is.Equal(rr.Code, http.StatusBadRequest)
}

func TestParsingMissingDatasetId(t *testing.T) {
	is := is.New(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should be stopped by the middleware anod it should not reach here")
	})

	testHandler := ParseDatasetIdMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusBadRequest)
}
