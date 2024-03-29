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
	router                  *mux.Router
	db                      *gorm.DB
	validate                *validator.Validate
	tokenAuth               *auth.TokenAuth
	userAuth                *auth.UserAuth
	authHandler             *handlers.AuthHandler
	datasetsHandler         *handlers.DatasetsHandler
	samplesHandler          *handlers.SamplesHandler
	usersHandler            *handlers.UsersHandler
	userDatasetPermsHandler *handlers.UserDatasetPermsHandler
}

func (a *App) Initialize() {
	db, err := utils.Init_db()
	if err != nil {
		log.Fatal(err)
	}

	if joinTableErr := db.SetupJoinTable(&models.User{}, "Datasets", &models.UserDataset{}); joinTableErr != nil {
		log.Fatal(joinTableErr)
	}

	a.validate = validator.New()

	a.db = db
	a.router = mux.NewRouter()

	a.tokenAuth = auth.NewTokenAuth(a.db)
	a.userAuth = auth.NewUserAuth(a.db)

	a.authHandler = handlers.NewAuthHandler(a.userAuth, a.tokenAuth)
	a.datasetsHandler = handlers.NewDatasetsHandler(db)
	a.samplesHandler = handlers.NewSamplesHandler(db)
	a.usersHandler = handlers.NewUsersHandler(db)
	a.userDatasetPermsHandler = handlers.NewUserDatasetPermsHandler(db)

	a.InitializeControllers()
}

func (a *App) InitializeControllers() {
	cors := mux_handlers.CORS(
		mux_handlers.AllowedHeaders([]string{"Content-Type", "sentry-trace", "baggage"}),
		mux_handlers.AllowedOrigins([]string{os.Getenv("ALLOWED_ORIGIN")}),
		mux_handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PATCH", "DELETE", "OPTIONS"}),
		mux_handlers.AllowCredentials(),
	)

	recoverHandler := mux_handlers.RecoveryHandler()
	sentryHandler := sentryhttp.New(sentryhttp.Options{Repanic: true})

	a.router.Use(recoverHandler, sentryHandler.Handle, middlewares.JSONResponseMiddleware, cors)

	authRouter := a.router.PathPrefix("/auth").Subrouter()
	authController := controllers.NewAuthController(a.tokenAuth, a.authHandler, a.validate)
	authController.Init(authRouter)

	userRouter := a.router.PathPrefix("/user").Subrouter()
	usersController := controllers.NewUsersController(a.tokenAuth)
	usersController.Init(userRouter)

	adminRouter := a.router.PathPrefix("/admin").Subrouter()
	adminController := controllers.NewAdminController(a.tokenAuth, a.usersHandler, a.userDatasetPermsHandler, a.validate)
	adminController.Init(adminRouter)

	datasetsRouter := a.router.PathPrefix("/datasets").Subrouter()
	datasetsController := controllers.NewDatasetsController(a.tokenAuth, a.datasetsHandler, a.samplesHandler, a.userDatasetPermsHandler, a.db)
	datasetsController.Init(datasetsRouter)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func (a *App) Run() {
	log.Fatal(http.ListenAndServe(":8010", logRequest(a.router)))
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
	sqlDB, err := a.db.DB()
	if err == nil {
		sqlDB.Close()
	}

	defer sqlDB.Close()
}
