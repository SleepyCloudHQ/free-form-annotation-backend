package middlewares

import (
	"backend/app/auth"
	"backend/app/models"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDBForIsAdminMiddlewareTests(t *testing.T) (*gorm.DB, func() error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}

	if migrationErr := db.AutoMigrate(&models.User{}); migrationErr != nil {
		t.Fatalf("failed to migrate user: %v", migrationErr)
	}

	sqlDB, sqlErr := db.DB()
	if sqlErr != nil {
		t.Fatalf("failed to obtain SQL DB: %v", sqlErr)
	}
	return db, sqlDB.Close
}

func TestIsAdminMiddlewareWithAdminRole(t *testing.T) {
	db, cleanup := setupDBForIsAdminMiddlewareTests(t)
	defer cleanup()

	is := is.New(t)
	userAuth := auth.NewUserAuth(db)

	user, userErr := userAuth.CreateUser("user@email.com", "pass", models.AdminRole)
	is.NoErr(userErr)

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
	db, cleanup := setupDBForIsAdminMiddlewareTests(t)
	defer cleanup()

	is := is.New(t)
	userAuth := auth.NewUserAuth(db)

	user, userErr := userAuth.CreateUser("user@email.com", "pass", models.AnnotatorRole)
	is.NoErr(userErr)

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
