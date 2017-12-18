package utils

import (
	"encoding/json"
	"net/http"

	"github.com/Mowinski/LastWatchedBackend/logger"
)

// RespondWithJSON prepare response as json with status code
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// ResponseBadRequestError return response with error and bad request status
func ResponseBadRequestError(w http.ResponseWriter, err error) {
	logger.Logger.Print("Error sent to client, error: ", err)
	RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
}
