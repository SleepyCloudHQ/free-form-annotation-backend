package controllers

import (
	"backend/app/auth"
	"backend/app/handlers"
	"backend/app/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/matryer/is"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDBForAuthControllerTests(t *testing.T) (*gorm.DB, func() error) {
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

func setupAuthController(t *testing.T) (*gorm.DB, func() error, *mux.Router) {
	db, cleanup := setupDBForAuthControllerTests(t)
	tokenAuth := auth.NewTokenAuth(db)
	userAuth := auth.NewUserAuth(db)
	authHandler := handlers.NewAuthHandler(userAuth, tokenAuth)
	validator := validator.New()
	router := mux.NewRouter()
	authController := NewAuthController(tokenAuth, authHandler, validator)
	authController.Init(router)
	return db, cleanup, router
}

func getCookieByName(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}

	return nil
}

func TestLogin(t *testing.T) {
	db, cleanup, router := setupAuthController(t)
	defer cleanup()
	is := is.New(t)

	email := "user@email.com"
	password := "pass"

	userAuth := auth.NewUserAuth(db)
	user, userErr := userAuth.CreateUser(email, password, models.AnnotatorRole)
	is.NoErr(userErr)

	requestBody := &LoginRequest{Email: email, Password: password}
	bodyBytes, marshalErr := json.Marshal(requestBody)
	is.NoErr(marshalErr)
	req := httptest.NewRequest("POST", "/login/", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusOK)

	// parse returned user
	responseUser := &models.User{}
	is.NoErr(json.NewDecoder(rr.Body).Decode(responseUser))
	is.Equal(responseUser.ID, user.ID)

	// check cookies
	authToken := &models.AuthToken{}
	cookies := rr.Result().Cookies()
	is.NoErr(db.Preload("RefreshToken").First(authToken, "user_id = ?", user.ID).Error)

	authCookie := getCookieByName(cookies, auth.AuthTokenCookieName)
	is.True(authCookie != nil)
	is.Equal(authCookie.Value, authToken.Token)

	refreshCookie := getCookieByName(cookies, auth.RefreshTokenCookieName)
	is.True(refreshCookie != nil)
	is.Equal(refreshCookie.Value, authToken.RefreshToken.Token)
}

func TestLoginWithWrongCredentials(t *testing.T) {
	db, cleanup, router := setupAuthController(t)
	defer cleanup()
	is := is.New(t)

	email := "user@email.com"
	password := "pass"

	userAuth := auth.NewUserAuth(db)
	_, userErr := userAuth.CreateUser(email, password, models.AnnotatorRole)
	is.NoErr(userErr)

	requestBody := &LoginRequest{Email: email, Password: "wrong pass"}
	bodyBytes, marshalErr := json.Marshal(requestBody)
	is.NoErr(marshalErr)
	req := httptest.NewRequest("POST", "/login/", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusBadRequest)
}

func TestLogoutWhileNotBeingLoggedIn(t *testing.T) {
	_, cleanup, router := setupAuthController(t)
	defer cleanup()
	is := is.New(t)

	req := httptest.NewRequest("POST", "/logout/", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusUnauthorized)
}

func TestLogout(t *testing.T) {
	db, cleanup, router := setupAuthController(t)
	defer cleanup()
	is := is.New(t)

	admin := models.User{Email: "user2", Role: models.AdminRole}
	is.NoErr(db.Create(&admin).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&admin)
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	req := httptest.NewRequest("POST", "/logout/", nil)
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusNoContent)

	// check cookies
	cookies := rr.Result().Cookies()
	responseAuthCookie := getCookieByName(cookies, auth.AuthTokenCookieName)
	is.Equal(responseAuthCookie.Value, "")
	is.Equal(responseAuthCookie.MaxAge, -1)

	responseRefreshCookie := getCookieByName(cookies, auth.RefreshTokenCookieName)
	is.Equal(responseRefreshCookie.Value, "")
	is.Equal(responseRefreshCookie.MaxAge, -1)
}

func TestRefreshTokenWithoutCookiesSet(t *testing.T) {
	_, cleanup, router := setupAuthController(t)
	defer cleanup()
	is := is.New(t)

	req := httptest.NewRequest("POST", "/refresh-token/", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusUnauthorized)
}

func TestRefreshToken(t *testing.T) {
	db, cleanup, router := setupAuthController(t)
	defer cleanup()
	is := is.New(t)

	admin := models.User{Email: "user2", Role: models.AdminRole}
	is.NoErr(db.Create(&admin).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&admin)
	is.NoErr(tokenErr)
	authCookie, refreshCookie := tokenAuth.CreateAuthCookies(authToken)

	req := httptest.NewRequest("POST", "/refresh-token/", nil)
	req.AddCookie(authCookie)
	req.AddCookie(refreshCookie)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusOK)

	// make sure cookies are not the same
	cookies := rr.Result().Cookies()
	responseAuthCookie := getCookieByName(cookies, auth.AuthTokenCookieName)
	is.True(responseAuthCookie != nil)
	is.True(responseAuthCookie.Value != authToken.Token)

	responseRefreshCookie := getCookieByName(cookies, auth.RefreshTokenCookieName)
	is.True(responseRefreshCookie != nil)
	is.True(responseRefreshCookie.Value != authToken.RefreshToken.Token)

	// cookies should have values from the new auth token (old one should be deleted)
	newAuthToken := &models.AuthToken{}
	is.NoErr(db.Preload("RefreshToken").First(newAuthToken, "user_id = ?", admin.ID).Error)
	is.True(newAuthToken.ID != authToken.ID)
	is.Equal(responseAuthCookie.Value, newAuthToken.Token)
	is.Equal(responseRefreshCookie.Value, newAuthToken.RefreshToken.Token)
}
