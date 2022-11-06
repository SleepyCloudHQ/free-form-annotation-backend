package controllers_utils

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"gorm.io/gorm"
)

func handleRecordNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	WriteError(errors.New("Not found"), w)
}

func HandleCommonErrors(err error, w http.ResponseWriter) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		handleRecordNotFound(w)
		return
	}

	log.Panic(err)
}

func WriteError(err error, w http.ResponseWriter) {
	errResponse := &ErrorResponse{
		ErrorMessage: err.Error(),
	}

	json.NewEncoder(w).Encode(errResponse)
}
