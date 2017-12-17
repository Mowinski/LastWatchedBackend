package utils

import (
	"encoding/json"
	"net/http"
)

// RespondWithJSON prepare response as json with status code
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
