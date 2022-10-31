package middlewares

import (
	"backend/app/handlers"
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

func setupDBForDatasetPermsMiddlewareTests(t *testing.T) (*gorm.DB, func() error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}

	if migrationErr := db.AutoMigrate(&models.UserDataset{}); migrationErr != nil {
		t.Fatalf("failed to migrate userdataset: %v", migrationErr)
	}

	sqlDB, sqlErr := db.DB()
	if sqlErr != nil {
		t.Fatalf("failed to obtain SQL DB: %v", sqlErr)
	}
	return db, sqlDB.Close
}

func TestAdminDatasetPerms(t *testing.T) {
	db, cleanup := setupDBForDatasetPermsMiddlewareTests(t)
	defer cleanup()

	is := is.New(t)
	admin := &models.User{Role: models.AdminRole}

	middleware := GetDatasetPermsMiddleware(handlers.NewUserDatasetPermsHandler(db))
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testHandler := middleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, admin)
	ctx = context.WithValue(ctx, DatasetIdContextKey, 0)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req.WithContext(ctx))

	is.Equal(rr.Code, http.StatusOK)
}

func TestUserWithoutDatasetPerms(t *testing.T) {
	db, cleanup := setupDBForDatasetPermsMiddlewareTests(t)
	defer cleanup()

	is := is.New(t)
	user := &models.User{Role: models.AnnotatorRole}

	middleware := GetDatasetPermsMiddleware(handlers.NewUserDatasetPermsHandler(db))
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should be stopped by the middleware anod it should not reach here")
	})

	testHandler := middleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, user)
	ctx = context.WithValue(ctx, DatasetIdContextKey, 0)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req.WithContext(ctx))

	is.Equal(rr.Code, http.StatusUnauthorized)
}

func TestUserWithDatasetPerms(t *testing.T) {
	db, cleanup := setupDBForDatasetPermsMiddlewareTests(t)
	defer cleanup()

	is := is.New(t)
	user := &models.User{Model: gorm.Model{ID: 0}, Role: models.AnnotatorRole}
	datasetId := 0

	handler := handlers.NewUserDatasetPermsHandler(db)
	handler.AddDatasetToUserPerms(user.ID, uint(datasetId))
	middleware := GetDatasetPermsMiddleware(handler)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testHandler := middleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, user)
	ctx = context.WithValue(ctx, DatasetIdContextKey, datasetId)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req.WithContext(ctx))

	is.Equal(rr.Code, http.StatusOK)
}

func TestMissingUserContext(t *testing.T) {
	db, cleanup := setupDBForDatasetPermsMiddlewareTests(t)
	defer cleanup()

	is := is.New(t)
	datasetId := 0

	handler := handlers.NewUserDatasetPermsHandler(db)
	middleware := GetDatasetPermsMiddleware(handler)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should be stopped by the middleware anod it should not reach here")
	})

	testHandler := middleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	ctx := context.WithValue(req.Context(), DatasetIdContextKey, datasetId)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req.WithContext(ctx))

	is.Equal(rr.Code, http.StatusInternalServerError)
}

func TestMissingDatasetContext(t *testing.T) {
	db, cleanup := setupDBForDatasetPermsMiddlewareTests(t)
	defer cleanup()

	is := is.New(t)
	user := &models.User{}

	handler := handlers.NewUserDatasetPermsHandler(db)
	middleware := GetDatasetPermsMiddleware(handler)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should be stopped by the middleware anod it should not reach here")
	})

	testHandler := middleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, user)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req.WithContext(ctx))

	is.Equal(rr.Code, http.StatusInternalServerError)
}
