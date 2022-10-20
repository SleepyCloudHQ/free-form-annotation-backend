package auth

import (
	utils "backend/app/controllers/utils"
	"backend/app/models"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"gorm.io/gorm"
)

var ErrTokenExpired = errors.New("token is expired")
var ErrInvalidToken = errors.New("invalid token provided")

const AuthTokenCookieName = "auth_token"
const RefreshTokenCookieName = "refresh_token"

type ContextKey string

const UserContextKey ContextKey = "user"

type TokenAuth struct {
	DB *gorm.DB
}

func NewTokenAuth(db *gorm.DB) *TokenAuth {
	return &TokenAuth{DB: db}
}

func (a *TokenAuth) CreateAuthToken(user *models.User) (*models.AuthToken, error) {
	randomAccessToken := make([]byte, 32)
	randomRefreshToken := make([]byte, 32)
	if _, accessTokenErr := rand.Read(randomAccessToken); accessTokenErr != nil {
		return nil, accessTokenErr
	}
	if _, refreshTokenErr := rand.Read(randomRefreshToken); refreshTokenErr != nil {
		return nil, refreshTokenErr
	}

	authToken := &models.AuthToken{
		Token:     base64.URLEncoding.EncodeToString(randomAccessToken),
		ExpiresAt: time.Now().Add(time.Minute * 60),
		RefreshToken: models.RefreshToken{
			Token:     base64.URLEncoding.EncodeToString(randomRefreshToken),
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
		},
		UserID: user.ID,
	}

	if result := a.DB.Create(authToken); result.Error != nil {
		return nil, result.Error
	}

	return authToken, nil
}

func (a *TokenAuth) CheckAuthToken(token string) (*models.User, error) {
	authToken := &models.AuthToken{}
	if result := a.DB.Preload("User").First(authToken, "token = ?", token); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, result.Error
	}

	if authToken.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return &authToken.User, nil
}

func (a *TokenAuth) RefreshToken(token string) (*models.AuthToken, error) {
	refreshToken := &models.RefreshToken{}
	if result := a.DB.First(refreshToken, "token = ?", token); result.Error != nil {
		return nil, result.Error
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	authToken := &models.AuthToken{}
	if authTokenResult := a.DB.Preload("User").First(&authToken, refreshToken.AuthTokenID); authTokenResult.Error != nil {
		return nil, authTokenResult.Error
	}

	// delete old auth and refersh tokens
	if result := a.DB.Delete(authToken); result.Error != nil {
		return nil, result.Error
	}

	if result := a.DB.Delete(refreshToken); result.Error != nil {
		return nil, result.Error
	}

	return a.CreateAuthToken(&authToken.User)
}

func (a *TokenAuth) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(AuthTokenCookieName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			utils.WriteError(errors.New("Unauthorized"), w)
			return
		}

		user, authErr := a.CheckAuthToken(cookie.Value)
		if authErr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			utils.WriteError(errors.New("Unauthorized"), w)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *TokenAuth) CreateAuthCookies(authToken *models.AuthToken) (*http.Cookie, *http.Cookie) {
	now := time.Now()
	authTokenCookie := &http.Cookie{
		Name:     AuthTokenCookieName,
		Value:    authToken.Token,
		Path:     "/",
		MaxAge:   int(authToken.ExpiresAt.Sub(now).Seconds()),
		Secure:   false,
		HttpOnly: true,
	}

	refreshTokenCookie := &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    authToken.RefreshToken.Token,
		Path:     "/",
		MaxAge:   int(authToken.RefreshToken.ExpiresAt.Sub(now).Seconds()),
		Secure:   false,
		HttpOnly: true,
	}

	return authTokenCookie, refreshTokenCookie
}

func (a *TokenAuth) CreateLogoutCookies() (*http.Cookie, *http.Cookie) {
	authTokenCookie := &http.Cookie{
		Name:     AuthTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   false,
		HttpOnly: true,
	}

	refreshTokenCookie := &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   false,
		HttpOnly: true,
	}

	return authTokenCookie, refreshTokenCookie
}
