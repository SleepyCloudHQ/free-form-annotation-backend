package controllers

import (
	"backend/app/auth"
	"backend/app/models"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type UsersController struct {
	tokenAuth *auth.TokenAuth
}

func NewUsersController(tokenAuth *auth.TokenAuth) *UsersController {
	return &UsersController{
		tokenAuth: tokenAuth,
	}
}

func (u *UsersController) Init(router *mux.Router) {
	router.Use(u.tokenAuth.AuthTokenMiddleware)
	router.HandleFunc("/", u.getUser).Methods("GET")
}

func (u *UsersController) getUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
