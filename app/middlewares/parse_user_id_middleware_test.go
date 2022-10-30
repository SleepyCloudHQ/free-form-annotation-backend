package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/matryer/is"
)

func TestParsingValidUserId(t *testing.T) {
	is := is.New(t)

	userId := 1

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextUserId, ok := r.Context().Value(UserIdContextKey).(int)
		is.True(ok)
		is.Equal(userId, contextUserId)
		w.WriteHeader(http.StatusOK)
	})

	testHandler := ParseUserIdMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)

	vars := map[string]string{
		UserIdVarKey: strconv.Itoa(userId),
	}

	testHandler.ServeHTTP(httptest.NewRecorder(), mux.SetURLVars(req, vars))
}

func TestParsingInvalidUserId(t *testing.T) {
	is := is.New(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should be stopped by the middleware anod it should not reach here")
	})

	testHandler := ParseUserIdMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)

	vars := map[string]string{
		UserIdVarKey: "this is not an id",
	}

	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, mux.SetURLVars(req, vars))
	is.Equal(rr.Code, http.StatusBadRequest)
}

func TestParsingMissingUserId(t *testing.T) {
	is := is.New(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should be stopped by the middleware anod it should not reach here")
	})

	testHandler := ParseUserIdMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusBadRequest)
}
