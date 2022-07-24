package main

import (
	"backend/app/auth"
	"backend/app/handlers"
	"backend/app/middlewares"
	"backend/app/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/go-playground/validator/v10"
	mux_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type App struct {
	Router                  *mux.Router
	DB                      *gorm.DB
	TokenAuth               *auth.TokenAuth
	UserAuth                *auth.UserAuth
	AuthHandler             *handlers.AuthHandler
	DatasetsHandler         *handlers.DatasetsHandler
	SamplesHandler          *handlers.SamplesHandler
	UsersHandler            *handlers.UsersHandler
	UserDatasetPermsHandler *handlers.UserDatasetPermsHandler
}

func (a *App) Initialize() {
	db, err := gorm.Open(sqlite.Open("test.db"))

	if err != nil {
		log.Fatal(err)
	}

	if jointTableErr := db.SetupJoinTable(&models.User{}, "Datasets", &models.UserDataset{}); jointTableErr != nil {
		log.Fatal(jointTableErr)
	}

	validate := validator.New()

	a.DB = db
	a.Router = mux.NewRouter()

	a.TokenAuth = auth.NewTokenAuth(a.DB)
	a.UserAuth = auth.NewUserAuth(a.DB)

	a.AuthHandler = handlers.NewAuthHandler(a.UserAuth, a.TokenAuth, validate)
	a.DatasetsHandler = handlers.NewDatasetsHandler(db)
	a.SamplesHandler = handlers.NewSamplesHandler(db)
	a.UsersHandler = handlers.NewUsersHandler(db, validate)
	a.UserDatasetPermsHandler = handlers.NewUserDatasetPermsHandler(db, validate)

	a.Migrate()
	a.InitializeRoutes()
}

func (a *App) Migrate() {
	//a.DB.AutoMigrate(&models.Dataset{})
	//a.DB.AutoMigrate(&models.Sample{})
}

func (a *App) InitializeRoutes() {
	cors := mux_handlers.CORS(
		mux_handlers.AllowedHeaders([]string{"content-type"}),
		mux_handlers.AllowedOrigins([]string{"http://localhost:3000"}),
		mux_handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PATCH", "DELETE"}),
		mux_handlers.AllowCredentials(),
	)

	a.Router.Use(cors, middlewares.JSONResponseMiddleware)

	authRouter := a.Router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/login/", a.login).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/refresh-token/", a.refreshToken).Methods("POST")
	authRouter.Handle("/logout/", a.TokenAuth.AuthTokenMiddleware(http.HandlerFunc(a.logout))).Methods("POST")

	userRouter := a.Router.PathPrefix("/user").Subrouter()
	userRouter.Use(a.TokenAuth.AuthTokenMiddleware)
	userRouter.HandleFunc("/", a.getUser).Methods("GET")

	adminRouter := a.Router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(a.TokenAuth.AuthTokenMiddleware, middlewares.IsAdminMiddleware)

	adminRouter.HandleFunc("/users/", a.getUsers).Methods("GET")

	adminUserManagementRouter := adminRouter.PathPrefix("/users/{userId:[0-9]+}").Subrouter()
	adminUserManagementRouter.Use(middlewares.ParseUserIdMiddleware)
	adminUserManagementRouter.HandleFunc("/roles/", a.patchUserRole).Methods("PATCH", "OPTIONS")
	adminUserManagementRouter.HandleFunc("/dataset-perms/", a.postUserDatasetPerm).Methods("POST")
	adminUserManagementRouter.HandleFunc("/dataset-perms/", a.deleteUserDatasetPerm).Methods("DELETE", "OPTIONS")

	datasetsRouter := a.Router.PathPrefix("/datasets").Subrouter()
	datasetsRouter.Use(a.TokenAuth.AuthTokenMiddleware)

	datasetsRouter.HandleFunc("/", a.getDatasets).Methods("GET")

	datasetRouter := datasetsRouter.PathPrefix("/{datasetId:[0-9]+}").Subrouter()
	datasetPermsMiddleware := middlewares.GetDatasetPermsMiddleware(a.DB)
	datasetRouter.Use(middlewares.ParseDatasetIdMiddleware, datasetPermsMiddleware)

	datasetRouter.HandleFunc("/", a.getDataset).Methods("GET")
	datasetRouter.HandleFunc("/samples/", a.getSamples).Methods("GET")
	datasetRouter.HandleFunc("/samples/next/", a.assignNextSample).Methods("GET")
	datasetRouter.HandleFunc("/samples/{status:[a-z]+}/", a.getSamplesWithStatus).Methods("GET")
	datasetRouter.HandleFunc("/samples/{sampleId:[0-9]+}/", a.getSample).Methods("GET")
	datasetRouter.HandleFunc("/samples/{sampleId:[0-9]+}/", a.patchSample).Methods("PATCH", "OPTIONS")
}

func (a *App) getUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a.UsersHandler.GetUsersWithDatasets())
}

func (a *App) patchUserRole(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middlewares.UserIdContextKey).(int)

	patchRoleRequest := &handlers.PatchUserRoleRequest{}
	if err := json.NewDecoder(r.Body).Decode(patchRoleRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		fmt.Println(err)
		return
	}

	user, patchErr := a.UsersHandler.PatchUserRole(uint(userId), patchRoleRequest)
	if patchErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(patchErr.Error()))
		fmt.Println(patchErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (a *App) postUserDatasetPerm(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middlewares.UserIdContextKey).(int)
	createUserDatasetPermRequest := &handlers.DatasetToUserPermsRequest{}
	if err := json.NewDecoder(r.Body).Decode(createUserDatasetPermRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		fmt.Println(err)
		return
	}

	createErr := a.UserDatasetPermsHandler.AddDatasetToUserPerms(uint(userId), createUserDatasetPermRequest)
	if createErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(createErr.Error()))
		fmt.Println(createErr)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *App) deleteUserDatasetPerm(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middlewares.UserIdContextKey).(int)
	deleteUserDatasetPermRequest := &handlers.DatasetToUserPermsRequest{}
	if err := json.NewDecoder(r.Body).Decode(deleteUserDatasetPermRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		fmt.Println(err)
		return
	}

	createErr := a.UserDatasetPermsHandler.DeleteDatasetToUserPerms(uint(userId), deleteUserDatasetPermRequest)
	if createErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(createErr.Error()))
		fmt.Println(createErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	loginRequest := &handlers.LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(loginRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		fmt.Println(err)
		return
	}
	loginResponse, loginErr := a.AuthHandler.Login(loginRequest)
	if loginErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(loginErr.Error()))
		fmt.Println(loginErr)
		return
	}
	http.SetCookie(w, loginResponse.Cookies.AuthTokenCookie)
	http.SetCookie(w, loginResponse.Cookies.RefreshTokenCookie)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loginResponse.User)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	loggedOutAuthTokenCookie, loggedOutRefreshTokenCookie := a.TokenAuth.CreateLogoutCookies()

	http.SetCookie(w, loggedOutAuthTokenCookie)
	http.SetCookie(w, loggedOutRefreshTokenCookie)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) refreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.RefreshTokenCookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	authCookies, loginErr := a.AuthHandler.RefreshToken(cookie.Value)
	if loginErr != nil {
		fmt.Println(loginErr)
		return
	}

	http.SetCookie(w, authCookies.AuthTokenCookie)
	http.SetCookie(w, authCookies.RefreshTokenCookie)
	w.WriteHeader(http.StatusOK)
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (a *App) getDatasets(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)

	var datasets *[]handlers.DatasetData
	if user.Role == models.AdminRole {
		datasets = a.DatasetsHandler.GetDatasets()
	} else {
		var datasetsErr error
		datasets, datasetsErr = a.DatasetsHandler.GetDatasetsForUser(user)
		if datasetsErr != nil {
			fmt.Println(datasetsErr)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datasets)
}

func (a *App) getDataset(w http.ResponseWriter, r *http.Request) {
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)

	dataset, datasetErr := a.DatasetsHandler.GetDataset(uint(datasetId))
	if datasetErr != nil {
		fmt.Println(datasetErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(datasetErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dataset)
}

func (a *App) getSamples(w http.ResponseWriter, r *http.Request) {
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)

	samples, samplesErr := a.SamplesHandler.GetSamples(uint(datasetId))
	if samplesErr != nil {
		fmt.Println(samplesErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(samplesErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(samples)
}

func (a *App) getSample(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)

	sampleIdString := vars["sampleId"]
	sampleId, err := strconv.Atoi(sampleIdString)
	if err != nil {
		fmt.Println("Error converting sample id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	sample, sampleErr := a.SamplesHandler.GetSample(uint(datasetId), uint(sampleId))
	if sampleErr != nil {
		fmt.Println(sampleErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(sampleErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sample)
}

func (a *App) getSamplesWithStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)

	// parse sample status
	statusString := vars["status"]
	status := models.StatusType(statusString)
	if statusErr := status.IsValid(); statusErr != nil {
		fmt.Println("Invalid status type")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(statusErr.Error()))
		return
	}

	samples, samplesErr := a.SamplesHandler.GetSamplesWithStatus(uint(datasetId), status)
	if samplesErr != nil {
		fmt.Println(samplesErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(samplesErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(samples)
}

func (a *App) assignNextSample(w http.ResponseWriter, r *http.Request) {
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)
	user := r.Context().Value(auth.UserContextKey).(*models.User)

	sample, sampleErr := a.SamplesHandler.AssignNextSample(uint(datasetId), user.ID)
	if sampleErr != nil {
		fmt.Println(sampleErr)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(sampleErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sample)
}

func (a *App) patchSample(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)

	sampleIdString := vars["sampleId"]
	sampleId, err := strconv.Atoi(sampleIdString)
	if err != nil {
		fmt.Println("Error converting sample id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	patchRequest := &handlers.PatchSampleRequest{}
	if err := json.NewDecoder(r.Body).Decode(patchRequest); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	sample, sampleErr := a.SamplesHandler.PatchSample(uint(datasetId), uint(sampleId), patchRequest)
	if sampleErr != nil {
		fmt.Println(sampleErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(sampleErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sample)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func (a *App) Run() {
	log.Fatal(http.ListenAndServe("localhost:8010", logRequest(a.Router)))
}

func checkLicence() {
	fmt.Println("checking licence")
	//os.Exit(1)
}

func main() {
	fmt.Println("Starting")
	s := gocron.NewScheduler(time.UTC)
	s.Every(5).Seconds().Do(checkLicence)
	s.StartAsync()

	a := App{}
	a.Initialize()
	a.Run()
	sqlDB, err := a.DB.DB()
	if err == nil {
		sqlDB.Close()
	}

	defer sqlDB.Close()
}
