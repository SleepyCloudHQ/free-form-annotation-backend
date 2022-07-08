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
	DatasetHandler *handlers.DatasetHandler
	SampleHandler  *handlers.SampleHandler
}

func (a *App) Initialize() {
	db, err := gorm.Open(sqlite.Open("test.db"))

	if err != nil {
		log.Fatal(err)
	}
	a.DB = db
	a.Router = mux.NewRouter()
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
		mux_handlers.AllowCredentials(),
	)

	a.Router.Use(cors, middlewares.JSONResponseMiddleware)

	a.Router.HandleFunc("/datasets", a.getDatasets).Methods("GET")
	a.Router.HandleFunc("/datasets/{id:[0-9]+}", a.getDataset).Methods("GET")
	a.Router.HandleFunc("/datasets/{id:[0-9]+}/samples", a.getSamples).Methods("GET")
	a.Router.HandleFunc("/datasets/{id:[0-9]+}/samples/next", a.assignNextSample).Methods("GET")
	a.Router.HandleFunc("/datasets/{id:[0-9]+}/samples/{status:[a-z]+}", a.getSamplesWithStatus).Methods("GET")
	a.Router.HandleFunc("/datasets/{id:[0-9]+}/samples/{sampleId:[0-9]+}", a.getSample).Methods("GET")
	a.Router.HandleFunc("/datasets/{id:[0-9]+}/samples/{sampleId:[0-9]+}", a.patchSample).Methods("PATCH")
}

func (a *App) getDatasets(w http.ResponseWriter, r *http.Request) {
	datasets := a.DatasetHandler.GetDatasets()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datasets)
}

func (a *App) getDataset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	datasetIdString := vars["id"]
	datasetId, err := strconv.Atoi(datasetIdString)
	if err != nil {
		fmt.Println("Error converting id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

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
	vars := mux.Vars(r)
	datasetIdString := vars["id"]
	datasetId, err := strconv.Atoi(datasetIdString)
	if err != nil {
		fmt.Println("Error converting id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

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
	datasetIdString := vars["id"]
	datasetId, datasetErr := strconv.Atoi(datasetIdString)
	if datasetErr != nil {
		fmt.Println("Error converting dataset id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(datasetErr.Error()))
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
	// parse dataset id
	datasetIdString := vars["id"]
	datasetId, err := strconv.Atoi(datasetIdString)
	if err != nil {
		fmt.Println("Error converting dataset id")
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
	vars := mux.Vars(r)
	datasetIdString := vars["id"]
	datasetId, err := strconv.Atoi(datasetIdString)
	if err != nil {
		fmt.Println("Error converting id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	sample, sampleErr := a.SampleHandler.AssignNextSample(uint(datasetId))
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
	datasetIdString := vars["id"]
	datasetId, datasetErr := strconv.Atoi(datasetIdString)
	if datasetErr != nil {
		fmt.Println("Error converting dataset id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(datasetErr.Error()))
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
