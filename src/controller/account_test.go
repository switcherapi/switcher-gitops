package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/model"
)

func TestAccountRegisterRoutes(t *testing.T) {
	// Create a router
	r := mux.NewRouter()

	// Test
	accountController.RegisterRoutes(r)

	// Assert
	assert.NotNil(t, r)
	assert.NotNil(t, r.GetRoute("GetAccount"))
	assert.NotNil(t, r.GetRoute("CreateAccount"))
	assert.NotNil(t, r.GetRoute("UpdateAccount"))
	assert.NotNil(t, r.GetRoute("DeleteAccount"))
}

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
	r := httptest.NewRequest(http.MethodPost, accountController.RouteAccountPath, nil)

	// Test
	accountController.CreateAccountHandler(w, r)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "{\"error\":\"Invalid request\"}", w.Body.String())
}

func TestCreateAccountHandlerErrorCreatingAccount(t *testing.T) {
	// Create an account
	accountController.CreateAccountHandler(givenAccountRequest(accountV1))

	// Create a request and response recorder
	w, r := givenAccountRequest(accountV1)

	// Test
	accountController.CreateAccountHandler(w, r)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "{\"error\":\"Error creating account\"}", w.Body.String())
}

func TestFetchAccountHandlerByDomainId(t *testing.T) {
	// Create an account
	accountController.CreateAccountHandler(givenAccountRequest(accountV1))

	// Create a request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, accountController.RouteAccountPath+"/123", nil)

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
	r := httptest.NewRequest(http.MethodGet, accountController.RouteAccountPath+"/111", nil)

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
	r := httptest.NewRequest(http.MethodPut, accountController.RouteAccountPath, nil)

	// Test
	accountController.UpdateAccountHandler(w, r)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "{\"error\":\"Invalid request\"}", w.Body.String())
}

func TestUpdateAccountHandlerErrorUpdatingAccount(t *testing.T) {
	// Create an account
	accountV1Copy := accountV1
	accountController.CreateAccountHandler(givenAccountRequest(accountV1Copy))

	// Replace the domain ID to force an error
	accountV1Copy.Domain.ID = "111"

	// Create a request and response recorder
	w, r := givenAccountRequest(accountV1Copy)

	// Test
	accountController.UpdateAccountHandler(w, r)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "{\"error\":\"Error updating account\"}", w.Body.String())
}

func TestDeleteAccountHandler(t *testing.T) {
	// Create an account
	accountController.CreateAccountHandler(givenAccountRequest(accountV1))

	// Create a request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, accountController.RouteAccountPath+"/123", nil)

	// Test
	accountController.DeleteAccountHandler(w, r)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteAccountHandlerNotFound(t *testing.T) {
	// Create a request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, accountController.RouteAccountPath+"/111", nil)

	// Test
	accountController.DeleteAccountHandler(w, r)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "{\"error\":\"Error deleting account: Account not found for domain.id: 111\"}", w.Body.String())
}

// Helpers

func givenAccountRequest(data model.Account) (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, accountController.RouteAccountPath, nil)

	// Encode the account request as JSON
	body, _ := json.Marshal(data)
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	return w, r
}
