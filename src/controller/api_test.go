package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckApiHandler(t *testing.T) {
	controller := ApiController{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/check", nil)

	controller.CheckApiHandler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"message":"API is working"}`, w.Body.String())
}
