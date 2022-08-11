package controllers_utils

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"gorm.io/gorm"
)

func handle_record_not_found(w http.ResponseWriter) {
	err := &ErrorResponse{
		ErrorMessage: "Not found",
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(err)
}

func Handle_common_errors(err error, w http.ResponseWriter) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		handle_record_not_found(w)
		return
	}

	log.Panic(err)
}
