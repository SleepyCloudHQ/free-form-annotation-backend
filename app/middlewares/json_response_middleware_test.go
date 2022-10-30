package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestContentType(t *testing.T) {
	is := is.New(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testHandler := JSONResponseMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	is.Equal(rr.Code, http.StatusOK)
	is.Equal(rr.Result().Header["Content-Type"][0], "application/json")
}
