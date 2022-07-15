package main

import (
	"backend/app/auth"
	"backend/app/handlers"
	"backend/app/middlewares"
	"backend/app/models"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	mux_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Router         *mux.Router
	DB             *gorm.DB
	TokenAuth      *auth.TokenAuth
	UserAuth       *auth.UserAuth
	AuthHandler    *handlers.AuthHandler
	DatasetHandler *handlers.DatasetHandler
	SampleHandler  *handlers.SampleHandler
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
	a.DatasetHandler = handlers.NewDatasetHandler(db)
	a.SampleHandler = handlers.NewSampleHandler(db)

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
		mux_handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PATCH"}),
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

	adminRouter.HandleFunc("/", a.getAdmin).Methods("GET")

	datasetsRouter := a.Router.PathPrefix("/datasets").Subrouter()
	datasetsRouter.Use(a.TokenAuth.AuthTokenMiddleware)

	datasetsRouter.HandleFunc("/", a.getDatasets).Methods("GET")

	datasetRouter := datasetsRouter.PathPrefix("/{dataset_id:[0-9]+}").Subrouter()
	datasetPermsMiddleware := middlewares.GetDatasetPermsMiddleware(a.DB)
	datasetRouter.Use(middlewares.ParseDatasetIdMiddleware, datasetPermsMiddleware)

	datasetRouter.HandleFunc("/", a.getDataset).Methods("GET")
	datasetRouter.HandleFunc("/samples/", a.getSamples).Methods("GET")
	datasetRouter.HandleFunc("/samples/next/", a.assignNextSample).Methods("GET")
	datasetRouter.HandleFunc("/samples/{status:[a-z]+}/", a.getSamplesWithStatus).Methods("GET")
	datasetRouter.HandleFunc("/samples/{sampleId:[0-9]+}/", a.getSample).Methods("GET")
	datasetRouter.HandleFunc("/samples/{sampleId:[0-9]+}/", a.patchSample).Methods("PATCH", "OPTIONS")

}

func (a *App) getAdmin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	//json.NewEncoder(w).Encode(loginResponse.User)
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
	datasets := a.DatasetHandler.GetDatasets()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datasets)
}

func (a *App) getDataset(w http.ResponseWriter, r *http.Request) {
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)

	dataset, datasetErr := a.DatasetHandler.GetDataset(uint(datasetId))
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

	samples, samplesErr := a.SampleHandler.GetSamples(uint(datasetId))
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

	sample, sampleErr := a.SampleHandler.GetSample(uint(datasetId), uint(sampleId))
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

	samples, samplesErr := a.SampleHandler.GetSamplesWithStatus(uint(datasetId), status)
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

	sample, sampleErr := a.SampleHandler.AssignNextSample(uint(datasetId))
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

	sample, sampleErr := a.SampleHandler.PatchSample(uint(datasetId), uint(sampleId), patchRequest)
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

func main() {
	fmt.Println("Starting")

	a := App{}
	a.Initialize()
	a.Run()
	sqlDB, err := a.DB.DB()
	if err == nil {
		sqlDB.Close()
	}

	defer sqlDB.Close()
}
