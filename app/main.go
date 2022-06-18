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
}

func (a *App) Initialize() {
	db, err := gorm.Open(sqlite.Open("test.db"))

	if err != nil {
		log.Fatal(err)
	}
	a.DB = db
	a.Router = mux.NewRouter()
	a.SessionHandler = handlers.NewSessionHandler(db)

	a.Migrate()
	a.InitializeRoutes()
}

func (a *App) Migrate() {
	a.DB.AutoMigrate(&models.Session{})
	a.DB.AutoMigrate(&models.Sample{})
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
	//a.Router.Handle("/posts/{id:[0-9]+}", authWrapper(a.updatePost)).Methods("PATCH")
	//a.Router.Handle("/posts/{id:[0-9]+}", authWrapper(a.deletePost)).Methods("DELETE")
	//a.Router.HandleFunc("/auth/login", a.login).Methods("POST", "OPTIONS")
	//a.Router.HandleFunc("/auth/register", a.register).Methods("POST")
	//a.Router.HandleFunc("/auth/refresh-token", a.refreshToken).Methods("POST")
	//a.Router.Handle("/user", authWrapper(a.getUser)).Methods("GET", "OPTIONS")
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
