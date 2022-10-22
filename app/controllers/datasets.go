package controllers

import (
	"backend/app/auth"
	utils "backend/app/controllers/utils"
	"backend/app/handlers"
	"backend/app/middlewares"
	"backend/app/models"
	dataset_export "backend/app/utils/dataset/export"
	dataset_import "backend/app/utils/dataset/import"
	"encoding/json"
	"errors"
	"fmt"
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
	authTokenMiddleware := middlewares.AuthTokenMiddleware(d.tokenAuth)
	router.Use(authTokenMiddleware)

	router.HandleFunc("/", d.getDatasets).Methods("GET", "OPTIONS")
	router.Handle("/", middlewares.IsAdminMiddleware(http.HandlerFunc(d.postDataset))).Methods("POST", "OPTIONS")

	datasetRouter := router.PathPrefix("/{datasetId:[0-9]+}").Subrouter()
	datasetPermsMiddleware := middlewares.GetDatasetPermsMiddleware(d.db)
	datasetRouter.Use(middlewares.ParseDatasetIdMiddleware, datasetPermsMiddleware)

	datasetRouter.HandleFunc("/", d.getDataset).Methods("GET", "OPTIONS")
	datasetRouter.Handle("/", middlewares.IsAdminMiddleware(http.HandlerFunc(d.deleteDataset))).Methods("DELETE", "OPTIONS")
	datasetRouter.Handle("/export/", middlewares.IsAdminMiddleware(http.HandlerFunc(d.exportDataset))).Methods("GET", "OPTIONS")
	datasetRouter.HandleFunc("/samples/", d.getSamples).Methods("GET", "OPTIONS")
	datasetRouter.HandleFunc("/samples/next/", d.assignNextSample).Methods("GET", "OPTIONS")
	datasetRouter.HandleFunc("/samples/{status:[a-z]+}/", d.getSamplesWithStatus).Methods("GET", "OPTIONS")
	datasetRouter.HandleFunc("/samples/{sampleId:[0-9]+}/", d.getSample).Methods("GET", "OPTIONS")
	datasetRouter.HandleFunc("/samples/{sampleId:[0-9]+}/", d.patchSample).Methods("PATCH", "OPTIONS")
}

func (d *DatasetsController) deleteDataset(w http.ResponseWriter, r *http.Request) {
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)
	if deleteErr := d.datasetsHandler.DeleteDataset(uint(datasetId)); deleteErr != nil {
		utils.HandleCommonErrors(deleteErr, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (d *DatasetsController) exportDataset(w http.ResponseWriter, r *http.Request) {
	datasetId := r.Context().Value(middlewares.DatasetIdContextKey).(int)

	dataset, datasetErr := d.datasetsHandler.GetDataset(uint(datasetId))
	if datasetErr != nil {
		utils.HandleCommonErrors(datasetErr, w)
		return
	}

	samples, samplesErr := d.samplesHandler.GetSamples(uint(datasetId))
	if samplesErr != nil {
		utils.HandleCommonErrors(samplesErr, w)
		return
	}

	jsonDataset, exportErr := dataset_export.ExportDataset(dataset, samples)
	if exportErr != nil {
		utils.HandleCommonErrors(exportErr, w)
		return
	}

	dispositionHeader := fmt.Sprintf("attachment; filename=dataset_%d.json", datasetId)
	w.Header().Set("Content-Disposition", dispositionHeader)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonDataset)
}

func (d *DatasetsController) getDatasets(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middlewares.UserContextKey).(*models.User)

	var datasets []*handlers.DatasetData
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
	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	jsonDataset, parsingErr := dataset_import.ParseDataset(file)
	if parsingErr != nil {
		log.Panic(parsingErr)
	}

	metadata, metadataErr := dataset_import.MarshalDatasetMetadata(jsonDataset.Metadata)
	if metadataErr != nil {
		log.Panic(metadataErr)
	}

	// create dataset
	dataset := &models.Dataset{
		Name:     jsonDataset.Name,
		Metadata: metadata,
	}

	if datasetCreateErr := d.db.Create(&dataset).Error; datasetCreateErr != nil {
		log.Fatal(datasetCreateErr)
	}

	samples, samplesErr := dataset_import.MapSampleDataToSample(jsonDataset.Samples, dataset.ID)
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

	dataset, datasetErr := d.datasetsHandler.GetDatasetData(uint(datasetId))
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
	user := r.Context().Value(middlewares.UserContextKey).(*models.User)

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

	// check whether status is a valid value
	if !patchRequest.Status.Valid {
		if statusErr := models.StatusType(patchRequest.Status.String).IsValid(); statusErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			utils.WriteError(statusErr, w)
		}
	}

	sample, sampleErr := d.samplesHandler.PatchSample(uint(datasetId), uint(sampleId), patchRequest)
	if sampleErr != nil {
		utils.HandleCommonErrors(sampleErr, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sample)
}
