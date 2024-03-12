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
	// Test
	payload, _ := json.Marshal(accountV1)
	req, _ := http.NewRequest(http.MethodPost, accountController.RouteAccountPath, bytes.NewBuffer(payload))
	response := executeRequest(req)

	// Assert
	var accountResponse model.Account
	err := json.NewDecoder(response.Body).Decode(&accountResponse)

	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Nil(t, err)
	assert.Equal(t, accountV1.Repository, accountResponse.Repository)
}

func TestCreateAccountHandlerInvalidRequest(t *testing.T) {
	// Test
	payload := []byte("")
	req, _ := http.NewRequest(http.MethodPost, accountController.RouteAccountPath, bytes.NewBuffer(payload))
	response := executeRequest(req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "{\"error\":\"Invalid request\"}", response.Body.String())
}

func TestCreateAccountHandlerErrorCreatingAccount(t *testing.T) {
	// Create an account
	accountController.CreateAccountHandler(givenAccountRequest(accountV1))

	// Test
	payload, _ := json.Marshal(accountV1)
	req, _ := http.NewRequest(http.MethodPost, accountController.RouteAccountPath, bytes.NewBuffer(payload))
	response := executeRequest(req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"error\":\"Error creating account\"}", response.Body.String())
}

func TestFetchAccountHandlerByDomainId(t *testing.T) {
	// Create an account
	accountController.CreateAccountHandler(givenAccountRequest(accountV1))

	// Test
	payload := []byte("")
	req, _ := http.NewRequest(http.MethodGet, accountController.RouteAccountPath+"/123", bytes.NewBuffer(payload))
	response := executeRequest(req)

	// Assert
	var accountResponse model.Account
	err := json.NewDecoder(response.Body).Decode(&accountResponse)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
	assert.Equal(t, accountV1.Repository, accountResponse.Repository)
}

func TestFetchAccountHandlerByDomainIdNotFound(t *testing.T) {
	// Test
	payload := []byte("")
	req, _ := http.NewRequest(http.MethodGet, accountController.RouteAccountPath+"/111", bytes.NewBuffer(payload))
	response := executeRequest(req)

	// Assert
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "{\"error\":\"Account not found\"}", response.Body.String())
}

func TestUpdateAccountHandler(t *testing.T) {
	// Create an account
	accountController.CreateAccountHandler(givenAccountRequest(accountV1))

	// Test
	payload, _ := json.Marshal(accountV2)
	req, _ := http.NewRequest(http.MethodPut, accountController.RouteAccountPath+"/123", bytes.NewBuffer(payload))
	response := executeRequest(req)

	// Assert
	var accountResponse model.Account
	err := json.NewDecoder(response.Body).Decode(&accountResponse)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
	assert.Equal(t, accountV2.Repository, accountResponse.Repository)
}

func TestUpdateAccountHandlerInvalidRequest(t *testing.T) {
	// Test
	payload := []byte("")
	req, _ := http.NewRequest(http.MethodPut, accountController.RouteAccountPath+"/123", bytes.NewBuffer(payload))
	response := executeRequest(req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "{\"error\":\"Invalid request\"}", response.Body.String())
}

func TestUpdateAccountHandlerErrorUpdatingAccount(t *testing.T) {
	// Create an account
	accountV1Copy := accountV1
	accountController.CreateAccountHandler(givenAccountRequest(accountV1Copy))

	// Replace the domain ID to force an error
	accountV1Copy.Domain.ID = "111"

	// Test
	payload, _ := json.Marshal(accountV1Copy)
	req, _ := http.NewRequest(http.MethodPut, accountController.RouteAccountPath+"/123", bytes.NewBuffer(payload))
	response := executeRequest(req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"error\":\"Error updating account\"}", response.Body.String())
}

func TestDeleteAccountHandler(t *testing.T) {
	// Create an account
	accountController.CreateAccountHandler(givenAccountRequest(accountV1))

	// Test
	req, _ := http.NewRequest(http.MethodDelete, accountController.RouteAccountPath+"/123", nil)
	response := executeRequest(req)

	// Assert
	assert.Equal(t, http.StatusNoContent, response.Code)
}

func TestDeleteAccountHandlerNotFound(t *testing.T) {
	// Test
	req, _ := http.NewRequest(http.MethodDelete, accountController.RouteAccountPath+"/111", nil)
	response := executeRequest(req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"error\":\"Error deleting account: Account not found for domain.id: 111\"}", response.Body.String())
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
