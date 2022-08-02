package controllers

import (
	"backend/app/auth"
	"backend/app/handlers"
	"backend/app/middlewares"
	"backend/app/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type DatasetsController struct {
	router          *mux.Router
	tokenAuth       *auth.TokenAuth
	datasetsHandler *handlers.DatasetsHandler
	samplesHandler  *handlers.SamplesHandler
	db              *gorm.DB
}

func NewDatasetsController(router *mux.Router, tokenAuth *auth.TokenAuth, datasetsHandler *handlers.DatasetsHandler, samplesHandler *handlers.SamplesHandler, db *gorm.DB) *DatasetsController {
	return &DatasetsController{
		router:          router,
		tokenAuth:       tokenAuth,
		datasetsHandler: datasetsHandler,
		samplesHandler:  samplesHandler,
		db:              db,
	}
}

func (d *DatasetsController) Init() {
	d.router.Use(d.tokenAuth.AuthTokenMiddleware)

	d.router.HandleFunc("/", d.getDatasets).Methods("GET")

	datasetRouter := d.router.PathPrefix("/{datasetId:[0-9]+}").Subrouter()
	datasetPermsMiddleware := middlewares.GetDatasetPermsMiddleware(d.db)
	datasetRouter.Use(middlewares.ParseDatasetIdMiddleware, datasetPermsMiddleware)

	datasetRouter.HandleFunc("/", d.getDataset).Methods("GET")
	datasetRouter.HandleFunc("/samples/", d.getSamples).Methods("GET")
	datasetRouter.HandleFunc("/samples/next/", d.assignNextSample).Methods("GET")
	datasetRouter.HandleFunc("/samples/{status:[a-z]+}/", d.getSamplesWithStatus).Methods("GET")
	datasetRouter.HandleFunc("/samples/{sampleId:[0-9]+}/", d.getSample).Methods("GET")
	datasetRouter.HandleFunc("/samples/{sampleId:[0-9]+}/", d.patchSample).Methods("PATCH", "OPTIONS")

}

func (d *DatasetsController) getDatasets(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)

	var datasets *[]handlers.DatasetData
	if user.Role == models.AdminRole {
		datasets = d.datasetsHandler.GetDatasets()
	} else {
		var datasetsErr error
		datasets, datasetsErr = d.datasetsHandler.GetDatasetsForUser(user)
		if datasetsErr != nil {
			fmt.Println(datasetsErr)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datasets)
}

func (d *DatasetsController) getDataset(w http.ResponseWriter, r *http.Request) {
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)

	dataset, datasetErr := d.datasetsHandler.GetDataset(uint(datasetId))
	if datasetErr != nil {
		fmt.Println(datasetErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(datasetErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dataset)
}

func (d *DatasetsController) getSamples(w http.ResponseWriter, r *http.Request) {
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)

	samples, samplesErr := d.samplesHandler.GetSamples(uint(datasetId))
	if samplesErr != nil {
		fmt.Println(samplesErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(samplesErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(samples)
}

func (d *DatasetsController) getSample(w http.ResponseWriter, r *http.Request) {
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

	sample, sampleErr := d.samplesHandler.GetSample(uint(datasetId), uint(sampleId))
	if sampleErr != nil {
		fmt.Println(sampleErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(sampleErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sample)
}

func (d *DatasetsController) getSamplesWithStatus(w http.ResponseWriter, r *http.Request) {
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

	samples, samplesErr := d.samplesHandler.GetSamplesWithStatus(uint(datasetId), status)
	if samplesErr != nil {
		fmt.Println(samplesErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(samplesErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(samples)
}

func (d *DatasetsController) assignNextSample(w http.ResponseWriter, r *http.Request) {
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)
	user := r.Context().Value(auth.UserContextKey).(*models.User)

	sample, sampleErr := d.samplesHandler.AssignNextSample(uint(datasetId), user.ID)
	if sampleErr != nil {
		fmt.Println(sampleErr)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(sampleErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sample)
}

func (d *DatasetsController) patchSample(w http.ResponseWriter, r *http.Request) {
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

	sample, sampleErr := d.samplesHandler.PatchSample(uint(datasetId), uint(sampleId), patchRequest)
	if sampleErr != nil {
		fmt.Println(sampleErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(sampleErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sample)
}
