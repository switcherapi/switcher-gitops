package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestApiRegisterRoutes(t *testing.T) {
	r := mux.NewRouter()
	apiController.RegisterRoutes(r)

	assert.NotNil(t, r)
	assert.NotNil(t, r.GetRoute("CheckApi"))
}

func TestCheckApiHandler(t *testing.T) {
	w, r := givenApiRequest()
	apiController.CheckApiHandler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"message":"API is working"}`, w.Body.String())
}

func givenApiRequest() (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/check", nil)

	return w, r
}
