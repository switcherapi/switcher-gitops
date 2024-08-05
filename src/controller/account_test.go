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
	t.Run("Should create an account", func(t *testing.T) {
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
	})

	t.Run("Should not create an account - invalid request", func(t *testing.T) {
		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodPost, accountController.RouteAccountPath, bytes.NewBuffer(payload))
		response := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.Equal(t, "{\"error\":\"Invalid request\"}", response.Body.String())
	})

	t.Run("Should not create an account - account already exists", func(t *testing.T) {
		// Create an account
		accountController.CreateAccountHandler(givenAccountRequest(accountV1))

		// Test
		payload, _ := json.Marshal(accountV1)
		req, _ := http.NewRequest(http.MethodPost, accountController.RouteAccountPath, bytes.NewBuffer(payload))
		response := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Equal(t, "{\"error\":\"Error creating account\"}", response.Body.String())
	})
}

func TestFetchAccountHandler(t *testing.T) {
	t.Run("Should fetch an account by domain ID", func(t *testing.T) {
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
	})

	t.Run("Should not fetch an account by domain ID - not found", func(t *testing.T) {
		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodGet, accountController.RouteAccountPath+"/111", bytes.NewBuffer(payload))
		response := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusNotFound, response.Code)
		assert.Equal(t, "{\"error\":\"Account not found\"}", response.Body.String())
	})
}

func TestUpdateAccountHandler(t *testing.T) {
	t.Run("Should update an account", func(t *testing.T) {
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
	})

	t.Run("Should not update an account - invalid request", func(t *testing.T) {
		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodPut, accountController.RouteAccountPath+"/123", bytes.NewBuffer(payload))
		response := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.Equal(t, "{\"error\":\"Invalid request\"}", response.Body.String())
	})

	t.Run("Should not update an account - not found", func(t *testing.T) {
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
	})
}

func TestDeleteAccountHandler(t *testing.T) {
	t.Run("Should delete an account", func(t *testing.T) {
		// Create an account
		accountController.CreateAccountHandler(givenAccountRequest(accountV1))

		// Test
		req, _ := http.NewRequest(http.MethodDelete, accountController.RouteAccountPath+"/123", nil)
		response := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("Should not delete an account - not found", func(t *testing.T) {
		// Test
		req, _ := http.NewRequest(http.MethodDelete, accountController.RouteAccountPath+"/111", nil)
		response := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Equal(t, "{\"error\":\"Error deleting account: Account not found for domain.id: 111\"}", response.Body.String())
	})
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
