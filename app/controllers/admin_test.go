package controllers

import (
	"backend/app/auth"
	"backend/app/handlers"
	"backend/app/models"
	"bytes"
	"encoding/json"
	"fmt"
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

func setupDBForAdminControllerTests(t *testing.T) (*gorm.DB, func() error) {
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

func setup(t *testing.T) (*gorm.DB, func() error, *mux.Router) {
	db, cleanup := setupDBForAdminControllerTests(t)
	tokenAuth := auth.NewTokenAuth(db)
	userHandler := handlers.NewUsersHandler(db)
	userDatasetPermsHandler := handlers.NewUserDatasetPermsHandler(db)
	validator := validator.New()
	router := mux.NewRouter()
	adminController := NewAdminController(tokenAuth, userHandler, userDatasetPermsHandler, validator)
	adminController.Init(router)
	return db, cleanup, router
}

func TestGetUsersWithoutAuth(t *testing.T) {
	_, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	// check auth
	req := httptest.NewRequest("GET", "/users/", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusUnauthorized)
}

func TestAdminGetUsers(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)
	tokenAuth := auth.NewTokenAuth(db)

	users := []models.User{
		{Email: "user1", Role: models.AdminRole},
		{Email: "user2", Role: models.AnnotatorRole},
	}
	is.NoErr(db.Create(&users).Error)

	authToken, tokenErr := tokenAuth.CreateAuthToken(&users[0])
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	req := httptest.NewRequest("GET", "/users/", nil)
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusOK)

	var usersFromResponse []models.User
	is.NoErr(json.NewDecoder(rr.Body).Decode(&usersFromResponse))
	is.Equal(usersFromResponse[0].ID, users[0].ID)
	is.Equal(usersFromResponse[1].ID, users[1].ID)
}

func TestAnnotatorGetUsers(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	tokenAuth := auth.NewTokenAuth(db)
	userAuth := auth.NewUserAuth(db)
	user, userErr := userAuth.CreateUser("admin@admin.com", "pass", models.AnnotatorRole)
	is.NoErr(userErr)

	authToken, tokenErr := tokenAuth.CreateAuthToken(user)
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	req := httptest.NewRequest("GET", "/users/", nil)
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusUnauthorized)
}

func TestPatchRolesWithoutAuth(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	users := []models.User{
		{Email: "user1", Role: models.AdminRole},
		{Email: "user2", Role: models.AnnotatorRole},
	}
	is.NoErr(db.Create(&users).Error)

	url := fmt.Sprintf("/users/%v/roles/", users[1].ID)
	requestBody := &PatchUserRoleRequest{Role: models.AdminRole}
	bodyBytes, marshalErr := json.Marshal(requestBody)
	is.NoErr(marshalErr)
	req := httptest.NewRequest("PATCH", url, bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusUnauthorized)

	refreshedUser := &models.User{}
	is.NoErr(db.First(refreshedUser, users[1].ID).Error)
	is.Equal(refreshedUser.Role, models.AnnotatorRole)
}

func TestPatchRolesNonExistingUser(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	admin := models.User{
		Email: "user1",
		Role:  models.AdminRole,
	}
	is.NoErr(db.Create(&admin).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&admin)
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	url := fmt.Sprintf("/users/%v/roles/", admin.ID+1)
	requestBody := &PatchUserRoleRequest{Role: models.AdminRole}
	bodyBytes, marshalErr := json.Marshal(requestBody)
	is.NoErr(marshalErr)
	req := httptest.NewRequest("PATCH", url, bytes.NewReader(bodyBytes))
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusBadRequest)
}

func TestPatchRolesAsAdmin(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	users := []models.User{
		{Email: "user1", Role: models.AdminRole},
		{Email: "user2", Role: models.AnnotatorRole},
	}
	is.NoErr(db.Create(&users).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&users[0])
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	url := fmt.Sprintf("/users/%v/roles/", users[1].ID)
	requestBody := &PatchUserRoleRequest{Role: models.AdminRole}
	bodyBytes, marshalErr := json.Marshal(requestBody)
	is.NoErr(marshalErr)

	req := httptest.NewRequest("PATCH", url, bytes.NewReader(bodyBytes))
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusOK)

	refreshedUser := &models.User{}
	is.NoErr(db.First(refreshedUser, users[1].ID).Error)
	is.Equal(refreshedUser.Role, models.AdminRole)
}

func TestPatchRolesAsAnnotator(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	annotator := models.User{
		Email: "user1",
		Role:  models.AnnotatorRole,
	}
	is.NoErr(db.Create(&annotator).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&annotator)
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	url := fmt.Sprintf("/users/%v/roles/", annotator.ID)
	req := httptest.NewRequest("PATCH", url, nil)
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusUnauthorized)

	refreshedUser := &models.User{}
	is.NoErr(db.First(refreshedUser, annotator.ID).Error)
	is.Equal(refreshedUser.Role, models.AnnotatorRole)
}

func TestPatchRolesInvalidRequest(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	users := []models.User{
		{Email: "user1", Role: models.AdminRole},
		{Email: "user2", Role: models.AnnotatorRole},
	}
	is.NoErr(db.Create(&users).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&users[0])
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	url := fmt.Sprintf("/users/%v/roles/", users[1].ID)
	requestBody := &PatchUserRoleRequest{Role: "invalid role"}
	bodyBytes, marshalErr := json.Marshal(requestBody)
	is.NoErr(marshalErr)

	req := httptest.NewRequest("PATCH", url, bytes.NewReader(bodyBytes))
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusBadRequest)

	refreshedUser := &models.User{}
	is.NoErr(db.First(refreshedUser, users[1].ID).Error)
	is.Equal(refreshedUser.Role, models.AnnotatorRole)
}

func TestPatchRolesInvalidRequestMissingRole(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	users := []models.User{
		{Email: "user1", Role: models.AdminRole},
		{Email: "user2", Role: models.AnnotatorRole},
	}
	is.NoErr(db.Create(&users).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&users[0])
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	url := fmt.Sprintf("/users/%v/roles/", users[1].ID)
	requestBody := &PatchUserRoleRequest{}
	bodyBytes, marshalErr := json.Marshal(requestBody)
	is.NoErr(marshalErr)

	req := httptest.NewRequest("PATCH", url, bytes.NewReader(bodyBytes))
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusBadRequest)

	refreshedUser := &models.User{}
	is.NoErr(db.First(refreshedUser, users[1].ID).Error)
	is.Equal(refreshedUser.Role, models.AnnotatorRole)
}

func TestPostUserDatasetPermWithoutAuth(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	users := []models.User{
		{Email: "user2", Role: models.AnnotatorRole},
	}
	is.NoErr(db.Create(&users).Error)

	url := fmt.Sprintf("/users/%v/dataset-perms/", users[0].ID)
	requestBody := &DatasetToUserPermsRequest{DatasetId: 0}
	bodyBytes, marshalErr := json.Marshal(requestBody)
	is.NoErr(marshalErr)
	req := httptest.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusUnauthorized)

	var count int64
	is.NoErr(db.Model(&models.UserDataset{}).Count(&count).Error)
	is.Equal(count, int64(0))
}

func TestPostUserDatasetPermAsAnnotator(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	annotator := models.User{Email: "user2", Role: models.AnnotatorRole}
	is.NoErr(db.Create(&annotator).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&annotator)
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	url := fmt.Sprintf("/users/%v/dataset-perms/", annotator.ID)
	requestBody := &DatasetToUserPermsRequest{DatasetId: 0}
	bodyBytes, marshalErr := json.Marshal(requestBody)

	is.NoErr(marshalErr)
	req := httptest.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusUnauthorized)

	var count int64
	is.NoErr(db.Model(&models.UserDataset{}).Count(&count).Error)
	is.Equal(count, int64(0))
}

func TestPostUserDatasetPermInvalidRequest(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	admin := models.User{Email: "user2", Role: models.AdminRole}
	is.NoErr(db.Create(&admin).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&admin)
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	url := fmt.Sprintf("/users/%v/dataset-perms/", admin.ID)

	// invalid request - no dataset id specified
	requestBody := &DatasetToUserPermsRequest{}
	bodyBytes, marshalErr := json.Marshal(requestBody)

	is.NoErr(marshalErr)
	req := httptest.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusBadRequest)

	var count int64
	is.NoErr(db.Model(&models.UserDataset{}).Count(&count).Error)
	is.Equal(count, int64(0))
}

func TestPostUserDatasetPerm(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	admin := models.User{Email: "user2", Role: models.AdminRole}
	is.NoErr(db.Create(&admin).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&admin)
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	url := fmt.Sprintf("/users/%v/dataset-perms/", admin.ID)
	requestBody := &DatasetToUserPermsRequest{DatasetId: 1}
	bodyBytes, marshalErr := json.Marshal(requestBody)

	is.NoErr(marshalErr)
	req := httptest.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusCreated)

	var count int64
	is.NoErr(db.Model(&models.UserDataset{}).Count(&count).Error)
	is.Equal(count, int64(1))
}

func TestDeleteUserDatasetPermWithoutAuth(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	users := []models.User{
		{Email: "user2", Role: models.AnnotatorRole},
	}
	is.NoErr(db.Create(&users).Error)

	url := fmt.Sprintf("/users/%v/dataset-perms/", users[0].ID)
	requestBody := &DatasetToUserPermsRequest{DatasetId: 0}
	bodyBytes, marshalErr := json.Marshal(requestBody)
	is.NoErr(marshalErr)
	req := httptest.NewRequest("DELETE", url, bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusUnauthorized)

	var count int64
	is.NoErr(db.Model(&models.UserDataset{}).Count(&count).Error)
	is.Equal(count, int64(0))
}

func TestDeleteUserDatasetPermAsAnnotator(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	annotator := models.User{Email: "user2", Role: models.AnnotatorRole}
	is.NoErr(db.Create(&annotator).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&annotator)
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	url := fmt.Sprintf("/users/%v/dataset-perms/", annotator.ID)
	requestBody := &DatasetToUserPermsRequest{DatasetId: 0}
	bodyBytes, marshalErr := json.Marshal(requestBody)

	is.NoErr(marshalErr)
	req := httptest.NewRequest("DELETE", url, bytes.NewReader(bodyBytes))
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusUnauthorized)

	var count int64
	is.NoErr(db.Model(&models.UserDataset{}).Count(&count).Error)
	is.Equal(count, int64(0))
}

func TestDeleteUserDatasetPermInvalidRequest(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	admin := models.User{Email: "user2", Role: models.AdminRole}
	is.NoErr(db.Create(&admin).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&admin)
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	url := fmt.Sprintf("/users/%v/dataset-perms/", admin.ID)

	// invalid request - no dataset id specified
	requestBody := &DatasetToUserPermsRequest{}
	bodyBytes, marshalErr := json.Marshal(requestBody)

	is.NoErr(marshalErr)
	req := httptest.NewRequest("DELETE", url, bytes.NewReader(bodyBytes))
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusBadRequest)

	var count int64
	is.NoErr(db.Model(&models.UserDataset{}).Count(&count).Error)
	is.Equal(count, int64(0))
}

func TestDeleteNonExistentUserDatasetPerm(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	admin := models.User{Email: "user2", Role: models.AdminRole}
	is.NoErr(db.Create(&admin).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&admin)
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	url := fmt.Sprintf("/users/%v/dataset-perms/", admin.ID)
	requestBody := &DatasetToUserPermsRequest{DatasetId: 1}
	bodyBytes, marshalErr := json.Marshal(requestBody)

	is.NoErr(marshalErr)
	req := httptest.NewRequest("DELETE", url, bytes.NewReader(bodyBytes))
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusNoContent)
}

func TestDeleteUserDatasetPerm(t *testing.T) {
	db, cleanup, router := setup(t)
	defer cleanup()
	is := is.New(t)

	admin := models.User{Email: "user2", Role: models.AdminRole}
	is.NoErr(db.Create(&admin).Error)

	perm := &models.UserDataset{UserID: admin.ID, DatasetID: 1}
	is.NoErr(db.Create(perm).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&admin)
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	url := fmt.Sprintf("/users/%v/dataset-perms/", admin.ID)
	requestBody := &DatasetToUserPermsRequest{DatasetId: perm.DatasetID}
	bodyBytes, marshalErr := json.Marshal(requestBody)

	is.NoErr(marshalErr)
	req := httptest.NewRequest("DELETE", url, bytes.NewReader(bodyBytes))
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusNoContent)

	var count int64
	is.NoErr(db.Model(&models.UserDataset{}).Count(&count).Error)
	is.Equal(count, int64(0))
}
