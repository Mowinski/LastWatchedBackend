package utils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Mowinski/LastWatchedBackend/logger"
)

func TestRespondWithJSON(t *testing.T) {
	w := httptest.NewRecorder()
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "OK"})

	if w.Result().Header.Get("Content-Type") != "application/json" {
		t.Errorf("Wrong Content-Type Header in response, got %s", w.Result().Header.Get("Content-Type"))
	}

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Wrong status code, got %d, expected 200", w.Result().StatusCode)
	}
}

func TestResponseBadRequestError(t *testing.T) {
	w := httptest.NewRecorder()

	logger.SetLogger("test_log_file.txt")
	defer os.Remove("test_log_file.txt")

	ResponseBadRequestError(w, fmt.Errorf("Test error"))

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Wrong status code, got %d, expected 400", w.Result().StatusCode)
	}

	if w.Body.String() != "{\"error\":\"Test error\"}" {
		t.Errorf("Wrong body, expected (%s), got (%s)", "{\"error\":\"Test error\"}", w.Body.String())
	}
}
