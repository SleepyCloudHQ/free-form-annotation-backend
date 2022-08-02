package main

import (
	"backend/app/auth"
	"backend/app/controllers"
	"backend/app/handlers"
	licence_checker "backend/app/licence"
	"backend/app/middlewares"
	"backend/app/models"
	"log"
	"net/http"
	"os"
	"time"

	"backend/app/utils"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-co-op/gocron"
	"github.com/go-playground/validator/v10"
	mux_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type App struct {
	Router                  *mux.Router
	DB                      *gorm.DB
	validate                *validator.Validate
	TokenAuth               *auth.TokenAuth
	UserAuth                *auth.UserAuth
	AuthHandler             *handlers.AuthHandler
	DatasetsHandler         *handlers.DatasetsHandler
	SamplesHandler          *handlers.SamplesHandler
	UsersHandler            *handlers.UsersHandler
	UserDatasetPermsHandler *handlers.UserDatasetPermsHandler
}

func (a *App) Initialize() {
	db, err := utils.Init_db()
	if err != nil {
		log.Fatal(err)
	}

	if jointTableErr := db.SetupJoinTable(&models.User{}, "Datasets", &models.UserDataset{}); jointTableErr != nil {
		log.Fatal(jointTableErr)
	}

	a.validate = validator.New()

	a.DB = db
	a.Router = mux.NewRouter()

	a.TokenAuth = auth.NewTokenAuth(a.DB)
	a.UserAuth = auth.NewUserAuth(a.DB)

	a.AuthHandler = handlers.NewAuthHandler(a.UserAuth, a.TokenAuth, a.validate)
	a.DatasetsHandler = handlers.NewDatasetsHandler(db)
	a.SamplesHandler = handlers.NewSamplesHandler(db)
	a.UsersHandler = handlers.NewUsersHandler(db, a.validate)
	a.UserDatasetPermsHandler = handlers.NewUserDatasetPermsHandler(db, a.validate)

	a.InitializeControllers()
}

func (a *App) InitializeControllers() {
	cors := mux_handlers.CORS(
		mux_handlers.AllowedHeaders([]string{"content-type"}),
		mux_handlers.AllowedOrigins([]string{os.Getenv("ALLOWED_ORIGIN")}),
		mux_handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PATCH", "DELETE"}),
		mux_handlers.AllowCredentials(),
	)

	recoverHandler := mux_handlers.RecoveryHandler()
	sentryHandler := sentryhttp.New(sentryhttp.Options{Repanic: true})

	a.Router.Use(recoverHandler, sentryHandler.Handle, cors, middlewares.JSONResponseMiddleware)

	authRouter := a.Router.PathPrefix("/auth").Subrouter()
	authController := controllers.NewAuthController(authRouter, a.TokenAuth, a.UserAuth, a.validate)
	authController.Init()

	userRouter := a.Router.PathPrefix("/user").Subrouter()
	usersController := controllers.NewUsersController(userRouter, a.TokenAuth)
	usersController.Init()

	adminRouter := a.Router.PathPrefix("/admin").Subrouter()
	adminController := controllers.NewAdminController(adminRouter, a.TokenAuth, a.UsersHandler, a.UserDatasetPermsHandler)
	adminController.Init()

	datasetsRouter := a.Router.PathPrefix("/datasets").Subrouter()
	datasetsController := controllers.NewDatasetsController(datasetsRouter, a.TokenAuth, a.DatasetsHandler, a.SamplesHandler, a.DB)
	datasetsController.Init()
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

func checkLicence(lc *licence_checker.LicenceChecker) func() {
	return func() {
		log.Println("Checking licence")
		licenceData, licenceErr := lc.CheckLicence()
		if licenceErr != nil {
			log.Fatal(licenceErr)
			os.Exit(1)
		}

		log.Printf("Licence for %s issued by %s is valid until %s.\n", licenceData.Subject, licenceData.Issuer, licenceData.ExpiresAt.String())
	}
}

func main() {
	log.Println("Starting")

	godotenv.Load()
	licenceFilePath := os.Getenv("LICENCE_FILE_PATH")

	licenceChecker, licenceCheckerErr := licence_checker.NewLicenceChecker(licenceFilePath)
	if licenceCheckerErr != nil {
		log.Fatal(licenceCheckerErr)
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DNS"),
		TracesSampleRate: 1.0,
	}); err != nil {
		log.Fatalf("Sentry initialization failed: %v\n", err)
	}

	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Day().Do(checkLicence(licenceChecker))
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
