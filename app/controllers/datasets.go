package controllers

import (
	"backend/app/auth"
	utils "backend/app/controllers/utils"
	"backend/app/handlers"
	"backend/app/middlewares"
	"backend/app/models"
	dataset_utils "backend/app/utils/dataset"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type DatasetsController struct {
	tokenAuth       *auth.TokenAuth
	datasetsHandler *handlers.DatasetsHandler
	samplesHandler  *handlers.SamplesHandler
	db              *gorm.DB
}

func NewDatasetsController(tokenAuth *auth.TokenAuth, datasetsHandler *handlers.DatasetsHandler, samplesHandler *handlers.SamplesHandler, db *gorm.DB) *DatasetsController {
	return &DatasetsController{
		tokenAuth:       tokenAuth,
		datasetsHandler: datasetsHandler,
		samplesHandler:  samplesHandler,
		db:              db,
	}
}

func (d *DatasetsController) Init(router *mux.Router) {
	router.Use(d.tokenAuth.AuthTokenMiddleware)

	router.HandleFunc("/", d.getDatasets).Methods("GET", "OPTIONS")
	router.HandleFunc("/", d.postDataset).Methods("POST", "OPTIONS")

	datasetRouter := router.PathPrefix("/{datasetId:[0-9]+}").Subrouter()
	datasetPermsMiddleware := middlewares.GetDatasetPermsMiddleware(d.db)
	datasetRouter.Use(middlewares.ParseDatasetIdMiddleware, datasetPermsMiddleware)

	datasetRouter.HandleFunc("/", d.getDataset).Methods("GET", "OPTIONS")
	datasetRouter.HandleFunc("/samples/", d.getSamples).Methods("GET", "OPTIONS")
	datasetRouter.HandleFunc("/samples/next/", d.assignNextSample).Methods("GET", "OPTIONS")
	datasetRouter.HandleFunc("/samples/{status:[a-z]+}/", d.getSamplesWithStatus).Methods("GET", "OPTIONS")
	datasetRouter.HandleFunc("/samples/{sampleId:[0-9]+}/", d.getSample).Methods("GET", "OPTIONS")
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
			utils.HandleCommonErrors(datasetsErr, w)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datasets)
}

func (d *DatasetsController) postDataset(w http.ResponseWriter, r *http.Request) {
	// user := r.Context().Value(auth.UserContextKey).(*models.User)

	r.ParseMultipartForm(32 << 20)
	datasetName := r.FormValue("datasetName")
	entityTags := dataset_utils.ParseTags(r.FormValue("entities"))
	relationshipTags := dataset_utils.ParseTags(r.FormValue("relationships"))

	if datasetName == "" {
		log.Panic("dataset name cannot be empty")
	}

	metadata, metadataErr := dataset_utils.CreateDatasetMetadata(entityTags, relationshipTags)
	if metadataErr != nil {
		log.Panic(metadataErr)
	}

	file, _, err := r.FormFile("datasetFile")
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	samplesData, samplesDataErr := dataset_utils.LoadSampleData(file)
	if samplesDataErr != nil {
		log.Panic(samplesDataErr)
	}

	// create dataset
	dataset := &models.Dataset{
		Name:     datasetName,
		Metadata: metadata,
	}

	if datasetCreateErr := d.db.Create(&dataset).Error; datasetCreateErr != nil {
		log.Fatal(datasetCreateErr)
	}

	samples, samplesErr := dataset_utils.MapSampleDataToSample(samplesData, dataset.ID)
	if samplesErr != nil {
		log.Panic(samplesErr)
	}

	// create samples in a batch
	if sampleCreateErr := d.db.Create(&samples).Error; sampleCreateErr != nil {
		log.Fatal(sampleCreateErr)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dataset)
}

func (d *DatasetsController) getDataset(w http.ResponseWriter, r *http.Request) {
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)

	dataset, datasetErr := d.datasetsHandler.GetDataset(uint(datasetId))
	if datasetErr != nil {
		utils.HandleCommonErrors(datasetErr, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dataset)
}

func (d *DatasetsController) getSamples(w http.ResponseWriter, r *http.Request) {
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)

	samples, samplesErr := d.samplesHandler.GetSamples(uint(datasetId))
	if samplesErr != nil {
		utils.HandleCommonErrors(samplesErr, w)
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
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(errors.New("Error converting sample id"), w)
		return
	}

	sample, sampleErr := d.samplesHandler.GetSample(uint(datasetId), uint(sampleId))
	if sampleErr != nil {
		utils.HandleCommonErrors(sampleErr, w)
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
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(errors.New("Invalid status type"), w)
		return
	}

	samples, samplesErr := d.samplesHandler.GetSamplesWithStatus(uint(datasetId), status)
	if samplesErr != nil {
		utils.HandleCommonErrors(samplesErr, w)
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
		utils.HandleCommonErrors(sampleErr, w)
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
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(errors.New("Error converting sample id"), w)
		return
	}

	patchRequest := &handlers.PatchSampleRequest{}
	if err := json.NewDecoder(r.Body).Decode(patchRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.WriteError(err, w)
		return
	}

	sample, sampleErr := d.samplesHandler.PatchSample(uint(datasetId), uint(sampleId), patchRequest)
	if sampleErr != nil {
		utils.HandleCommonErrors(sampleErr, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sample)
}
