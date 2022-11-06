package middlewares

import (
	"backend/app/models"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestIsAdminMiddlewareWithAdminRole(t *testing.T) {
	is := is.New(t)
	user := &models.User{Role: models.AdminRole}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testHandler := IsAdminMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, user)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req.WithContext(ctx))

	is.Equal(rr.Code, http.StatusOK)
}

func TestIsAdminMiddlewareWithAnnotatorRole(t *testing.T) {
	is := is.New(t)
	user := &models.User{Role: models.AnnotatorRole}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should be stopped by the middleware anod it should not reach here")
	})

	testHandler := IsAdminMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, user)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req.WithContext(ctx))

	is.Equal(rr.Code, http.StatusUnauthorized)
}

func TestIsAdminMiddlewareWithoutUser(t *testing.T) {
	is := is.New(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should be stopped by the middleware anod it should not reach here")
	})

	testHandler := IsAdminMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	is.Equal(rr.Code, http.StatusUnauthorized)
}
