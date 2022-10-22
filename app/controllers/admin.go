package controllers

import (
	"backend/app/auth"
	"backend/app/handlers"
	"backend/app/middlewares"
	"encoding/json"
	"net/http"

	utils "backend/app/controllers/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type DatasetToUserPermsRequest struct {
	DatasetId uint `json:"dataset_id" validate:"required"`
}

type AdminController struct {
	tokenAuth               *auth.TokenAuth
	usersHandler            *handlers.UsersHandler
	userDatasetPermsHandler *handlers.UserDatasetPermsHandler
	Validator               *validator.Validate
}

func NewAdminController(tokenAuth *auth.TokenAuth, usersHandler *handlers.UsersHandler, userDatasetPermsHandler *handlers.UserDatasetPermsHandler, validator *validator.Validate) *AdminController {
	return &AdminController{
		tokenAuth:               tokenAuth,
		usersHandler:            usersHandler,
		userDatasetPermsHandler: userDatasetPermsHandler,
		Validator:               validator,
	}
}

func (a *AdminController) Init(router *mux.Router) {
	authTokenMiddleware := middlewares.AuthTokenMiddleware(a.tokenAuth)
	router.Use(authTokenMiddleware, middlewares.IsAdminMiddleware)
	router.HandleFunc("/users/", a.getUsers).Methods("GET", "OPTIONS")

	adminUserManagementRouter := router.PathPrefix("/users/{userId:[0-9]+}").Subrouter()
	adminUserManagementRouter.Use(middlewares.ParseUserIdMiddleware)
	adminUserManagementRouter.HandleFunc("/roles/", a.patchUserRole).Methods("PATCH", "OPTIONS")
	adminUserManagementRouter.HandleFunc("/dataset-perms/", a.postUserDatasetPerm).Methods("POST", "OPTIONS")
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
		utils.WriteError(err, w)
		return
	}

	user, patchErr := a.usersHandler.PatchUserRole(uint(userId), patchRoleRequest)
	if patchErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(patchErr, w)
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
		utils.WriteError(err, w)
		return
	}

	if valErr := a.Validator.Struct(createUserDatasetPermRequest); valErr != nil {
		utils.HandleCommonErrors(valErr.(validator.ValidationErrors), w)
		return
	}

	createErr := a.userDatasetPermsHandler.AddDatasetToUserPerms(uint(userId), createUserDatasetPermRequest.DatasetId)
	if createErr != nil {
		utils.HandleCommonErrors(createErr, w)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *AdminController) deleteUserDatasetPerm(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middlewares.UserIdContextKey).(int)
	deleteUserDatasetPermRequest := &handlers.DatasetToUserPermsRequest{}
	if err := json.NewDecoder(r.Body).Decode(deleteUserDatasetPermRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(err, w)
		return
	}

	if valErr := a.Validator.Struct(deleteUserDatasetPermRequest); valErr != nil {
		utils.HandleCommonErrors(valErr.(validator.ValidationErrors), w)
		return
	}

	deleteErr := a.userDatasetPermsHandler.DeleteDatasetToUserPerms(uint(userId), deleteUserDatasetPermRequest.DatasetId)
	if deleteErr != nil {
		utils.HandleCommonErrors(deleteErr, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
