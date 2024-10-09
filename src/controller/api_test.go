package controller

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckApiHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/check", nil)
	response := executeRequest(req, r, "")

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), "All good")
}

func TestApiDocsHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/docs", nil)
	response := executeRequest(req, r, "")
	assert.Equal(t, http.StatusOK, response.Code)
}
