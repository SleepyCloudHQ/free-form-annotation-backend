package main

import (
	"backend/app/handlers"
	"backend/app/middlewares"
	"backend/app/models"
	"encoding/json"
	"fmt"
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
	SessionHandler *handlers.SessionHandler
	SampleHandler  *handlers.SampleHandler
}

func (a *App) Initialize() {
	db, err := gorm.Open(sqlite.Open("test.db"))

	if err != nil {
		log.Fatal(err)
	}
	a.DB = db
	a.Router = mux.NewRouter()
	a.SessionHandler = handlers.NewSessionHandler(db)
	a.SampleHandler = handlers.NewSampleHandler(db)

	a.Migrate()
	a.InitializeRoutes()
}

func (a *App) Migrate() {
	//a.DB.AutoMigrate(&models.Session{})
	//a.DB.AutoMigrate(&models.Sample{})
}

func (a *App) InitializeRoutes() {
	cors := mux_handlers.CORS(
		mux_handlers.AllowedHeaders([]string{"content-type"}),
		mux_handlers.AllowedOrigins([]string{"http://localhost:3000"}),
		mux_handlers.AllowCredentials(),
	)

	a.Router.Use(cors, middlewares.JSONResponseMiddleware)

	a.Router.HandleFunc("/sessions", a.getSessions).Methods("GET")
	a.Router.HandleFunc("/sessions/{id:[0-9]+}", a.getSession).Methods("GET")
	a.Router.HandleFunc("/sessions/{id:[0-9]+}/samples", a.getSamples).Methods("GET")
	a.Router.HandleFunc("/sessions/{id:[0-9]+}/samples/next", a.assignNextSample).Methods("GET")
	a.Router.HandleFunc("/sessions/{id:[0-9]+}/samples/{status:[a-z]+}", a.getSamplesWithStatus).Methods("GET")
	a.Router.HandleFunc("/sessions/{id:[0-9]+}/samples/{sampleId:[0-9]+}", a.getSample).Methods("GET")
	a.Router.HandleFunc("/sessions/{id:[0-9]+}/samples/{sampleId:[0-9]+}", a.patchSample).Methods("PATCH")
}

func (a *App) getSessions(w http.ResponseWriter, r *http.Request) {
	sessions := a.SessionHandler.GetSessions()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sessions)
}

func (a *App) getSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionIdString := vars["id"]
	sessionId, err := strconv.Atoi(sessionIdString)
	if err != nil {
		fmt.Println("Error converting id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	session, sessionErr := a.SessionHandler.GetSession(uint(sessionId))
	if sessionErr != nil {
		fmt.Println(sessionErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(sessionErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(session)
}

func (a *App) getSamples(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionIdString := vars["id"]
	sessionId, err := strconv.Atoi(sessionIdString)
	if err != nil {
		fmt.Println("Error converting id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	samples, samplesErr := a.SampleHandler.GetSamples(uint(sessionId))
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
	sessionIdString := vars["id"]
	sessionId, sessionErr := strconv.Atoi(sessionIdString)
	if sessionErr != nil {
		fmt.Println("Error converting session id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(sessionErr.Error()))
		return
	}

	sampleIdString := vars["sampleId"]
	sampleId, err := strconv.Atoi(sampleIdString)
	if err != nil {
		fmt.Println("Error converting sample id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	sample, sampleErr := a.SampleHandler.GetSample(uint(sessionId), uint(sampleId))
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
	// parse session id
	sessionIdString := vars["id"]
	sessionId, err := strconv.Atoi(sessionIdString)
	if err != nil {
		fmt.Println("Error converting session id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// parse sample status
	statusString := vars["status"]
	status := models.StatusType(statusString)
	if statusErr := status.IsValid(); statusErr != nil {
		fmt.Println("Invalid status type")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(statusErr.Error()))
		return
	}

	samples, samplesErr := a.SampleHandler.GetSamplesWithStatus(uint(sessionId), status)
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
	vars := mux.Vars(r)
	sessionIdString := vars["id"]
	sessionId, err := strconv.Atoi(sessionIdString)
	if err != nil {
		fmt.Println("Error converting id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	sample, sampleErr := a.SampleHandler.AssignNextSample(uint(sessionId))
	if sampleErr != nil {
		fmt.Println(sampleErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(sampleErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sample)
}

func (a *App) patchSample(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionIdString := vars["id"]
	sessionId, sessionErr := strconv.Atoi(sessionIdString)
	if sessionErr != nil {
		fmt.Println("Error converting session id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(sessionErr.Error()))
		return
	}

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

	sample, sampleErr := a.SampleHandler.PatchSample(uint(sessionId), uint(sampleId), patchRequest)
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
	log.Fatal(http.ListenAndServe(":8010", logRequest(a.Router)))
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
