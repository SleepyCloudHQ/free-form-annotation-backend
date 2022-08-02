package controllers

import (
	"backend/app/auth"
	"backend/app/models"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type UsersController struct {
	router    *mux.Router
	tokenAuth *auth.TokenAuth
}

func NewUsersController(router *mux.Router, tokenAuth *auth.TokenAuth) *UsersController {
	return &UsersController{
		router:    router,
		tokenAuth: tokenAuth,
	}
}

func (u *UsersController) InitPaths() {
	u.router.Use(u.tokenAuth.AuthTokenMiddleware)
	u.router.HandleFunc("/", u.getUser).Methods("GET")
}

func (u *UsersController) getUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
