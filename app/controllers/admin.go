package controllers

import (
	"backend/app/auth"
	"backend/app/handlers"
	"backend/app/middlewares"
	"encoding/json"
	"net/http"

	utils "backend/app/controllers/utils"
	"github.com/gorilla/mux"
)

type AdminController struct {
	tokenAuth               *auth.TokenAuth
	usersHandler            *handlers.UsersHandler
	userDatasetPermsHandler *handlers.UserDatasetPermsHandler
}

func NewAdminController(tokenAuth *auth.TokenAuth, usersHandler *handlers.UsersHandler, userDatasetPermsHandler *handlers.UserDatasetPermsHandler) *AdminController {
	return &AdminController{
		tokenAuth:               tokenAuth,
		usersHandler:            usersHandler,
		userDatasetPermsHandler: userDatasetPermsHandler,
	}
}

func (a *AdminController) Init(router *mux.Router) {
	router.Use(a.tokenAuth.AuthTokenMiddleware, middlewares.IsAdminMiddleware)
	router.HandleFunc("/users/", a.getUsers).Methods("GET")

	adminUserManagementRouter := router.PathPrefix("/users/{userId:[0-9]+}").Subrouter()
	adminUserManagementRouter.Use(middlewares.ParseUserIdMiddleware)
	adminUserManagementRouter.HandleFunc("/roles/", a.patchUserRole).Methods("PATCH", "OPTIONS")
	adminUserManagementRouter.HandleFunc("/dataset-perms/", a.postUserDatasetPerm).Methods("POST")
	adminUserManagementRouter.HandleFunc("/dataset-perms/", a.deleteUserDatasetPerm).Methods("DELETE", "OPTIONS")
}

func (a *AdminController) getUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a.usersHandler.GetUsersWithDatasets())
}

func (a *AdminController) patchUserRole(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middlewares.UserIdContextKey).(int)

	patchRoleRequest := &handlers.PatchUserRoleRequest{}
	if err := json.NewDecoder(r.Body).Decode(patchRoleRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	user, patchErr := a.usersHandler.PatchUserRole(uint(userId), patchRoleRequest)
	if patchErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(patchErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (a *AdminController) postUserDatasetPerm(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middlewares.UserIdContextKey).(int)
	createUserDatasetPermRequest := &handlers.DatasetToUserPermsRequest{}
	if err := json.NewDecoder(r.Body).Decode(createUserDatasetPermRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	createErr := a.userDatasetPermsHandler.AddDatasetToUserPerms(uint(userId), createUserDatasetPermRequest)
	if createErr != nil {
		utils.Handle_common_errors(createErr, w)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *AdminController) deleteUserDatasetPerm(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middlewares.UserIdContextKey).(int)
	deleteUserDatasetPermRequest := &handlers.DatasetToUserPermsRequest{}
	if err := json.NewDecoder(r.Body).Decode(deleteUserDatasetPermRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	deleteErr := a.userDatasetPermsHandler.DeleteDatasetToUserPerms(uint(userId), deleteUserDatasetPermRequest)
	if deleteErr != nil {
		utils.Handle_common_errors(deleteErr, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
