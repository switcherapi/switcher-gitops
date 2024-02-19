package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/model"
)

func TestCreateAccountHandler(t *testing.T) {
	// Create a request and response recorder
	w, r := givenAccountRequest(accountV1)

	// Test
	accountController.CreateAccountHandler(w, r)

	// Assert
	var accountResponse model.Account
	err := json.NewDecoder(w.Body).Decode(&accountResponse)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Nil(t, err)
	assert.Equal(t, accountV1.Repository, accountResponse.Repository)
}

func TestCreateAccountHandlerInvalidRequest(t *testing.T) {
	// Create a request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", accountController.RouteAccountPath, nil)

	// Test
	accountController.CreateAccountHandler(w, r)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "{\"error\":\"Invalid request\"}", w.Body.String())
}

func TestFetchAccountHandlerByDomainId(t *testing.T) {
	// Create an account
	accountController.CreateAccountHandler(givenAccountRequest(accountV1))

	// Create a request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", accountController.RouteAccountPath+"/123", nil)

	// Test
	accountController.FetchAccountHandler(w, r)

	// Assert
	var accountResponse model.Account
	err := json.NewDecoder(w.Body).Decode(&accountResponse)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Nil(t, err)
	assert.Equal(t, accountV1.Repository, accountResponse.Repository)
}

func TestFetchAccountHandlerByDomainIdNotFound(t *testing.T) {
	// Create a request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", accountController.RouteAccountPath+"/111", nil)

	// Test
	accountController.FetchAccountHandler(w, r)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "{\"error\":\"Account not found\"}", w.Body.String())
}

func TestUpdateAccountHandler(t *testing.T) {
	// Create an account
	accountController.CreateAccountHandler(givenAccountRequest(accountV1))

	// Create a request and response recorder
	w, r := givenAccountRequest(accountV2)

	// Test
	accountController.UpdateAccountHandler(w, r)

	// Assert
	var accountResponse model.Account
	err := json.NewDecoder(w.Body).Decode(&accountResponse)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Nil(t, err)
	assert.Equal(t, accountV2.Repository, accountResponse.Repository)
}

func TestUpdateAccountHandlerInvalidRequest(t *testing.T) {
	// Create a request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", accountController.RouteAccountPath, nil)

	// Test
	accountController.UpdateAccountHandler(w, r)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "{\"error\":\"Invalid request\"}", w.Body.String())
}

// Helpers

func givenAccountRequest(data model.Account) (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", accountController.RouteAccountPath, nil)

	// Encode the account request as JSON
	body, _ := json.Marshal(data)
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	return w, r
}
