package controllers

import (
	"backend/app/auth"
	utils "backend/app/controllers/utils"
	"backend/app/handlers"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type RegisterRequest struct {
	Email            string `json:"email" validate:"required"`
	Password         string `json:"password" validate:"required,eqfield=PasswordRepeated"`
	PasswordRepeated string `json:"password_repeated" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthController struct {
	authHandler *handlers.AuthHandler
	tokenAuth   *auth.TokenAuth
	validator   *validator.Validate
}

func NewAuthController(tokenAuth *auth.TokenAuth, authHandler *handlers.AuthHandler, validator *validator.Validate) *AuthController {
	controller := &AuthController{
		tokenAuth:   tokenAuth,
		authHandler: authHandler,
		validator:   validator,
	}

	return controller
}

func (a *AuthController) Init(router *mux.Router) {
	router.HandleFunc("/login/", a.login).Methods("POST", "OPTIONS")
	router.HandleFunc("/refresh-token/", a.refreshToken).Methods("POST", "OPTIONS")
	router.Handle("/logout/", a.tokenAuth.AuthTokenMiddleware(http.HandlerFunc(a.logout))).Methods("POST", "OPTIONS")
}

func (a *AuthController) login(w http.ResponseWriter, r *http.Request) {
	loginRequest := &LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(loginRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(err, w)
		return
	}

	if valErr := a.validator.Struct(loginRequest); valErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(valErr.(validator.ValidationErrors), w)
		return
	}

	user, authCookies, loginErr := a.authHandler.Login(loginRequest.Email, loginRequest.Password)
	if loginErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(loginErr, w)
		return
	}

	http.SetCookie(w, authCookies.AuthTokenCookie)
	http.SetCookie(w, authCookies.RefreshTokenCookie)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (a *AuthController) register(w http.ResponseWriter, r *http.Request) {
	registerRequest := &RegisterRequest{}
	if err := json.NewDecoder(r.Body).Decode(registerRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(err, w)
		return
	}

	if valErr := a.validator.Struct(registerRequest); valErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(valErr.(validator.ValidationErrors), w)
		return
	}

	user, authCookies, registerErr := a.authHandler.Register(registerRequest.Email, registerRequest.Password)
	if registerErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(registerErr, w)
		return
	}

	http.SetCookie(w, authCookies.AuthTokenCookie)
	http.SetCookie(w, authCookies.RefreshTokenCookie)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
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
