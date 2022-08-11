package controllers

import (
	"backend/app/auth"
	utils "backend/app/controllers/utils"
	"backend/app/handlers"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type AuthController struct {
	authHandler *handlers.AuthHandler
	tokenAuth   *auth.TokenAuth
}

func NewAuthController(tokenAuth *auth.TokenAuth, authHandler *handlers.AuthHandler) *AuthController {
	controller := &AuthController{
		tokenAuth:   tokenAuth,
		authHandler: authHandler,
	}

	return controller
}

func (a *AuthController) Init(router *mux.Router) {
	router.HandleFunc("/login/", a.login).Methods("POST", "OPTIONS")
	router.HandleFunc("/refresh-token/", a.refreshToken).Methods("POST")
	router.Handle("/logout/", a.tokenAuth.AuthTokenMiddleware(http.HandlerFunc(a.logout))).Methods("POST")
}

func (a *AuthController) login(w http.ResponseWriter, r *http.Request) {
	loginRequest := &handlers.LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(loginRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(err, w)
		return
	}
	loginResponse, loginErr := a.authHandler.Login(loginRequest)
	if loginErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(loginErr, w)
		return
	}

	http.SetCookie(w, loginResponse.Cookies.AuthTokenCookie)
	http.SetCookie(w, loginResponse.Cookies.RefreshTokenCookie)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loginResponse.User)
}

func (a *AuthController) logout(w http.ResponseWriter, r *http.Request) {
	loggedOutAuthTokenCookie, loggedOutRefreshTokenCookie := a.tokenAuth.CreateLogoutCookies()

	http.SetCookie(w, loggedOutAuthTokenCookie)
	http.SetCookie(w, loggedOutRefreshTokenCookie)
	w.WriteHeader(http.StatusNoContent)
}

func (a *AuthController) refreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.RefreshTokenCookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		utils.WriteError(errors.New("Unauthorized"), w)
		return
	}

	authCookies, loginErr := a.authHandler.RefreshToken(cookie.Value)
	if loginErr != nil {
		log.Panic(loginErr)
		return
	}

	http.SetCookie(w, authCookies.AuthTokenCookie)
	http.SetCookie(w, authCookies.RefreshTokenCookie)
	w.WriteHeader(http.StatusOK)
}
