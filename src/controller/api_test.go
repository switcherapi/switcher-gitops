package controller

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckApiHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/check", nil)
	response := executeRequest(req)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "{\"message\":\"API is working\"}", response.Body.String())
}
