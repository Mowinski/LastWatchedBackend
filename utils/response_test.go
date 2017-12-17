package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
