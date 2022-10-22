package handlers

import (
	"backend/app/auth"
	"backend/app/models"
	"net/http"
)

type AuthCookies struct {
	AuthTokenCookie    *http.Cookie
	RefreshTokenCookie *http.Cookie
}

type AuthHandler struct {
	userAuth  *auth.UserAuth
	tokenAuth *auth.TokenAuth
}

func NewAuthHandler(userAuth *auth.UserAuth, tokenAuth *auth.TokenAuth) *AuthHandler {
	return &AuthHandler{
		userAuth:  userAuth,
		tokenAuth: tokenAuth,
	}
}

func (a *AuthHandler) createAuthCookiesForUser(user *models.User) (*AuthCookies, error) {
	authToken, authTokenErr := a.tokenAuth.CreateAuthToken(user)
	if authTokenErr != nil {
		return nil, authTokenErr
	}
	authTokenCookie, refreshTokenCookie := a.tokenAuth.CreateAuthCookies(authToken)
	return &AuthCookies{
		AuthTokenCookie:    authTokenCookie,
		RefreshTokenCookie: refreshTokenCookie,
	}, nil
}

func (a *AuthHandler) Register(email string, password string) (*models.User, *AuthCookies, error) {
	user, userErr := a.userAuth.CreateUser(email, password, models.AnnotatorRole)
	if userErr != nil {
		return nil, nil, userErr
	}

	authCookies, authCookiesErr := a.createAuthCookiesForUser(user)
	if authCookiesErr != nil {
		return nil, nil, authCookiesErr
	}

	return user, authCookies, nil
}

func (a *AuthHandler) Login(email string, password string) (*models.User, *AuthCookies, error) {
	user, checkUserErr := a.userAuth.CheckUserPassword(email, password)
	if checkUserErr != nil {
		return nil, nil, checkUserErr
	}

	authCookies, authCookiesErr := a.createAuthCookiesForUser(user)
	if authCookiesErr != nil {
		return nil, nil, authCookiesErr
	}

	return user, authCookies, nil
}

func (a *AuthHandler) RefreshToken(refreshToken string) (*AuthCookies, error) {
	authToken, authTokenErr := a.tokenAuth.RefreshToken(refreshToken)
	if authTokenErr != nil {
		return nil, authTokenErr
	}

	authTokenCookie, refreshTokenCookie := a.tokenAuth.CreateAuthCookies(authToken)
	return &AuthCookies{
		AuthTokenCookie:    authTokenCookie,
		RefreshTokenCookie: refreshTokenCookie,
	}, nil
}
