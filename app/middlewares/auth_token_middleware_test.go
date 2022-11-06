package middlewares

import (
	"backend/app/auth"
	"backend/app/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matryer/is"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDBForAuthTokenMiddlewareTests(t *testing.T) (*gorm.DB, func() error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}

	if migrationErr := db.AutoMigrate(&models.User{}); migrationErr != nil {
		t.Fatalf("failed to migrate user: %v", migrationErr)
	}

	if migrationErr := db.AutoMigrate(&models.AuthToken{}); migrationErr != nil {
		t.Fatalf("failed to migrate auth token: %v", migrationErr)
	}

	if migrationErr := db.AutoMigrate(&models.RefreshToken{}); migrationErr != nil {
		t.Fatalf("failed to migrate refresh token: %v", migrationErr)
	}

	sqlDB, sqlErr := db.DB()
	if sqlErr != nil {
		t.Fatalf("failed to obtain SQL DB: %v", sqlErr)
	}
	return db, sqlDB.Close
}

func TestMiddlewareWithProperAuthToken(t *testing.T) {
	db, cleanup := setupDBForAuthTokenMiddlewareTests(t)
	defer cleanup()

	is := is.New(t)

	tokenAuth := auth.NewTokenAuth(db)
	userAuth := auth.NewUserAuth(db)

	user, userErr := userAuth.CreateUser("user@email.com", "pass", models.AdminRole)
	is.NoErr(userErr)
	authToken, tokenErr := tokenAuth.CreateAuthToken(user)
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	middleware := AuthTokenMiddleware(tokenAuth)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextUser, ok := r.Context().Value(UserContextKey).(*models.User)
		is.True(ok)
		is.True(contextUser != nil)
		is.Equal(contextUser.ID, user.ID)
	})

	testHandler := middleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	req.AddCookie(authCookie)
	testHandler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestMiddlewareWithMissingCookie(t *testing.T) {
	db, cleanup := setupDBForAuthTokenMiddlewareTests(t)
	defer cleanup()

	is := is.New(t)

	tokenAuth := auth.NewTokenAuth(db)

	middleware := AuthTokenMiddleware(tokenAuth)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should be stopped by the middleware anod it should not reach here")
	})

	testHandler := middleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	is.Equal(rr.Code, http.StatusUnauthorized)
}

func TestMiddlewareWithInvalidAuthToken(t *testing.T) {
	db, cleanup := setupDBForAuthTokenMiddlewareTests(t)
	defer cleanup()

	is := is.New(t)

	tokenAuth := auth.NewTokenAuth(db)
	authTokenCookie := &http.Cookie{
		Name:   auth.AuthTokenCookieName,
		Value:  "invalid token",
		MaxAge: int(time.Hour),
	}

	middleware := AuthTokenMiddleware(tokenAuth)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should be stopped by the middleware anod it should not reach here")
	})

	testHandler := middleware(nextHandler)
	req := httptest.NewRequest("GET", "http://testing", nil)
	req.AddCookie(authTokenCookie)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	is.Equal(rr.Code, http.StatusUnauthorized)
}
