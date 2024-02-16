package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "Some message"}
	status := http.StatusOK

	ResponseJSON(w, data, status)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Some message", response["message"])
}
