package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

const NOT_FOUND = "/not-found"

func TestCreateAccountHandler(t *testing.T) {
	t.Run("Should create an account", func(t *testing.T) {
		// Test
		accountV1Copy := &accountV1
		accountV1Copy.Domain.ID = "123-controller-create-account"
		payload, _ := json.Marshal(accountV1Copy)
		req, _ := http.NewRequest(http.MethodPost, accountController.RouteAccountPath, bytes.NewBuffer(payload))
		response := executeRequest(req)

		// Assert
		var accountResponse model.Account
		err := json.NewDecoder(response.Body).Decode(&accountResponse)

		assert.Equal(t, http.StatusCreated, response.Code)
		assert.Nil(t, err)
		assert.Equal(t, accountV1Copy.Repository, accountResponse.Repository)

		token, _ := utils.Decrypt(accountResponse.Token, config.GetEnv("GIT_TOKEN_PRIVATE_KEY"))
		assert.Equal(t, accountV1Copy.Token, token)
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
		accountV1Copy := &accountV1
		accountV1Copy.Domain.ID = "123-controller-fetch-account"
		accountController.CreateAccountHandler(givenAccountRequest(*accountV1Copy))

		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodGet, accountController.RouteAccountPath+"/"+accountV1Copy.Domain.ID, bytes.NewBuffer(payload))
		response := executeRequest(req)

		// Assert
		var accountResponse model.Account
		err := json.NewDecoder(response.Body).Decode(&accountResponse)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Nil(t, err)
		assert.Equal(t, accountV1Copy.Repository, accountResponse.Repository)
	})

	t.Run("Should not fetch an account by domain ID - not found", func(t *testing.T) {
		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodGet, accountController.RouteAccountPath+NOT_FOUND, bytes.NewBuffer(payload))
		response := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusNotFound, response.Code)
		assert.Equal(t, "{\"error\":\"Account not found\"}", response.Body.String())
	})
}

func TestUpdateAccountHandler(t *testing.T) {
	t.Run("Should update an account", func(t *testing.T) {
		// Create an account
		accountV1Copy := &accountV1
		accountV1Copy.Domain.ID = "123-controller-update-account"
		accountController.CreateAccountHandler(givenAccountRequest(*accountV1Copy))

		// Test
		accountV2Copy := &accountV2
		accountV2Copy.Domain.ID = accountV1Copy.Domain.ID
		accountV2Copy.Domain.Message = "Updated successfully"
		payload, _ := json.Marshal(accountV2Copy)
		req, _ := http.NewRequest(http.MethodPut, accountController.RouteAccountPath+"/"+accountV2Copy.Domain.ID, bytes.NewBuffer(payload))
		response := executeRequest(req)

		// Assert
		var accountResponse model.Account
		err := json.NewDecoder(response.Body).Decode(&accountResponse)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Nil(t, err)
		assert.NotEmpty(t, accountResponse.Token)
		assert.Equal(t, accountV1.Repository, accountResponse.Repository)
		assert.Equal(t, model.StatusSynced, accountResponse.Domain.Status)
		assert.Equal(t, "Updated successfully", accountResponse.Domain.Message)
		assert.Equal(t, "5m", accountResponse.Settings.Window)
	})

	t.Run("Should update account token only", func(t *testing.T) {
		// Create an account
		accountV1Copy := &accountV1
		accountV1Copy.Domain.ID = "123-controller-update-account-token"
		accountController.CreateAccountHandler(givenAccountRequest(*accountV1Copy))

		// Test
		accountRequest := *accountV1Copy
		accountRequest.Token = "new-token"

		payload, _ := json.Marshal(accountRequest)
		req, _ := http.NewRequest(http.MethodPut, accountController.RouteAccountPath+"/"+accountV1Copy.Domain.ID, bytes.NewBuffer(payload))
		response := executeRequest(req)

		// Assert
		var accountResponse model.Account
		err := json.NewDecoder(response.Body).Decode(&accountResponse)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Nil(t, err)
		assert.Equal(t, accountV1Copy.Repository, accountResponse.Repository)

		encryptedToken := utils.Encrypt(accountV1Copy.Token, config.GetEnv("GIT_TOKEN_PRIVATE_KEY"))
		assert.NotEqual(t, encryptedToken, accountResponse.Token)

		decryptedToken, _ := utils.Decrypt(accountResponse.Token, config.GetEnv("GIT_TOKEN_PRIVATE_KEY"))
		assert.Equal(t, "new-token", decryptedToken)
	})

	t.Run("Should not update an account - invalid request", func(t *testing.T) {
		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodPut, accountController.RouteAccountPath+"/invalid-request", bytes.NewBuffer(payload))
		response := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.Equal(t, "{\"error\":\"Invalid request\"}", response.Body.String())
	})

	t.Run("Should not update an account - not found", func(t *testing.T) {
		// Create an account
		accountV1Copy := &accountV1
		accountV1Copy.Domain.ID = "123-controller-update-account-not-found"
		accountController.CreateAccountHandler(givenAccountRequest(*accountV1Copy))

		// Replace the domain ID to force an error
		accountV1Copy.Domain.ID = "111"

		// Test
		payload, _ := json.Marshal(accountV1Copy)
		req, _ := http.NewRequest(http.MethodPut, accountController.RouteAccountPath+NOT_FOUND, bytes.NewBuffer(payload))
		response := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Equal(t, "{\"error\":\"Error updating account\"}", response.Body.String())
	})
}

func TestDeleteAccountHandler(t *testing.T) {
	t.Run("Should delete an account", func(t *testing.T) {
		// Create an account
		accountV1Copy := &accountV1
		accountV1Copy.Domain.ID = "123-controller-delete-account"
		accountController.CreateAccountHandler(givenAccountRequest(*accountV1Copy))

		// Test
		req, _ := http.NewRequest(http.MethodDelete, accountController.RouteAccountPath+"/"+accountV1Copy.Domain.ID, nil)
		response := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("Should not delete an account - not found", func(t *testing.T) {
		// Test
		req, _ := http.NewRequest(http.MethodDelete, accountController.RouteAccountPath+NOT_FOUND, nil)
		response := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Equal(t, "{\"error\":\"Error deleting account: Account not found for domain.id: not-found\"}", response.Body.String())
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
