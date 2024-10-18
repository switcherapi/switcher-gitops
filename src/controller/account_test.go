package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/repository"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

func TestToken(t *testing.T) {
	token := generateToken("test", time.Minute)
	assert.NotEmpty(t, token)
}

func TestCreateAccountHandler(t *testing.T) {
	token := generateToken("test", time.Minute)

	t.Run("Should create an account", func(t *testing.T) {
		// Test
		accountV1.Domain.ID = "123-controller-create-account"
		payload, _ := json.Marshal(accountV1)
		req, _ := http.NewRequest(http.MethodPost, accountController.routeAccountPath, bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		var accountResponse model.Account
		err := json.NewDecoder(response.Body).Decode(&accountResponse)

		assert.Equal(t, http.StatusCreated, response.Code)
		assert.Nil(t, err)
		assert.Equal(t, accountV1.Repository, accountResponse.Repository)
		assert.Contains(t, accountResponse.Token, "...")
	})

	t.Run("Should not create an account - invalid request", func(t *testing.T) {
		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodPost, accountController.routeAccountPath, bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.Equal(t, "{\"error\":\"Invalid request\"}", response.Body.String())
	})

	t.Run("Should not create an account - account already exists", func(t *testing.T) {
		// Create an account
		accountController.CreateAccountHandler(givenAccountRequest(accountV1))

		// Test
		payload, _ := json.Marshal(accountV1)
		req, _ := http.NewRequest(http.MethodPost, accountController.routeAccountPath, bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Equal(t, "{\"error\":\"Error creating account\"}", response.Body.String())
	})
}

func TestFetchAccountHandler(t *testing.T) {
	token := generateToken("test", time.Minute)

	t.Run("Should fetch an account by domain ID / environment", func(t *testing.T) {
		// Create an account
		accountV1.Domain.ID = "123-controller-fetch-account"
		accountController.CreateAccountHandler(givenAccountRequest(accountV1))

		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodGet, accountController.routeAccountPath+"/"+accountV1.Domain.ID+"/"+accountV1.Environment, bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		var accountResponse model.Account
		err := json.NewDecoder(response.Body).Decode(&accountResponse)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Nil(t, err)
		assert.Equal(t, accountV1.Repository, accountResponse.Repository)
		assert.Contains(t, accountResponse.Token, "...")
	})

	t.Run("Should not fetch an account by domain ID / environment - not found", func(t *testing.T) {
		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodGet, accountController.routeAccountPath+"/not-found/default", bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		assert.Equal(t, http.StatusNotFound, response.Code)
		assert.Equal(t, "{\"error\":\"Account not found\"}", response.Body.String())
	})

	t.Run("Should fetch all accounts by domain ID", func(t *testing.T) {
		// Create an account
		accountV1.Domain.ID = "123-controller-fetch-all-accounts"
		accountController.CreateAccountHandler(givenAccountRequest(accountV1))
		accountV1.Environment = "staging"
		accountController.CreateAccountHandler(givenAccountRequest(accountV1))

		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodGet, accountController.routeAccountPath+"/"+accountV1.Domain.ID, bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		var accountsResponse []model.Account
		err := json.NewDecoder(response.Body).Decode(&accountsResponse)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(accountsResponse))
		assert.Contains(t, accountsResponse[0].Token, "...")
		assert.Contains(t, accountsResponse[1].Token, "...")
	})

	t.Run("Should not fetch all accounts by domain ID - not found", func(t *testing.T) {
		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodGet, accountController.routeAccountPath+"/not-found", bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		assert.Equal(t, http.StatusNotFound, response.Code)
		assert.Equal(t, "{\"error\":\"Not found accounts for domain: not-found\"}", response.Body.String())
	})
}

func TestUpdateAccountHandler(t *testing.T) {
	token := generateToken("test", time.Minute)

	t.Run("Should update an account", func(t *testing.T) {
		// Create an account
		accountV1.Domain.ID = "123-controller-update-account"
		accountV1.Environment = "default"
		accountController.CreateAccountHandler(givenAccountRequest(accountV1))

		// Update the account
		accountV2.Domain.ID = "123-controller-update-account"
		accountV2.Environment = "default"
		accountV2.Domain.Message = "Updated successfully"

		// Test
		payload, _ := json.Marshal(accountV2)
		req, _ := http.NewRequest(http.MethodPut, accountController.routeAccountPath, bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

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
		accountV1.Domain.ID = "123-controller-update-account-token"
		accountController.CreateAccountHandler(givenAccountRequest(accountV1))

		// Test
		accountRequest := accountV1
		accountRequest.Token = "new-token"

		payload, _ := json.Marshal(accountRequest)
		req, _ := http.NewRequest(http.MethodPut, accountController.routeAccountPath, bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		var accountResponse model.Account
		err := json.NewDecoder(response.Body).Decode(&accountResponse)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Nil(t, err)
		assert.Equal(t, accountV1.Repository, accountResponse.Repository)
		assert.Contains(t, accountResponse.Token, "...")

		accountRepository := repository.NewAccountRepositoryMongo(mongoDb)
		accountFromDb, _ := accountRepository.FetchByAccountId(accountResponse.ID.Hex())

		decryptedToken, _ := utils.Decrypt(accountFromDb.Token, config.GetEnv("GIT_TOKEN_PRIVATE_KEY"))
		assert.Equal(t, "new-token", decryptedToken)
	})

	t.Run("Should not update an account - invalid request", func(t *testing.T) {
		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodPut, accountController.routeAccountPath, bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.Equal(t, "{\"error\":\"Invalid request\"}", response.Body.String())
	})

	t.Run("Should not update an account - not found", func(t *testing.T) {
		// Create an account
		accountV1.Domain.ID = "123-controller-update-account-not-found"
		accountController.CreateAccountHandler(givenAccountRequest(accountV1))

		// Replace the domain ID to force an error
		accountV1.Domain.ID = "111"

		// Test
		payload, _ := json.Marshal(accountV1)
		req, _ := http.NewRequest(http.MethodPut, accountController.routeAccountPath, bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Equal(t, "{\"error\":\"Error updating account\"}", response.Body.String())
	})
}

func TestUpdateAccountTokensHandler(t *testing.T) {
	token := generateToken("test", time.Minute)

	t.Run("Should update account tokens", func(t *testing.T) {
		// Create accounts
		account1 := accountV1
		account1.Domain.ID = "123-controller-update-account-tokens"
		account1.Environment = "default"
		accountController.CreateAccountHandler(givenAccountRequest(account1))

		account2 := accountV1
		account2.Domain.ID = "123-controller-update-account-tokens"
		account2.Environment = "staging"
		accountController.CreateAccountHandler(givenAccountRequest(account2))

		// Test
		payload, _ := json.Marshal(AccountTokensRequest{
			DomainId:     account1.Domain.ID,
			Environments: []string{account1.Environment, account2.Environment},
			Token:        "new-token",
		})

		req, _ := http.NewRequest(http.MethodPut, accountController.routeAccountPath+"/"+account1.Domain.ID, bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		var accountTokensResponse AccountTokensResponse
		err := json.NewDecoder(response.Body).Decode(&accountTokensResponse)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Nil(t, err)

		accountRepository := repository.NewAccountRepositoryMongo(mongoDb)
		accountFromDb1 := accountRepository.FetchAllByDomainId(account1.Domain.ID)

		decryptedToken1, _ := utils.Decrypt(accountFromDb1[0].Token, config.GetEnv("GIT_TOKEN_PRIVATE_KEY"))
		decryptedToken2, _ := utils.Decrypt(accountFromDb1[1].Token, config.GetEnv("GIT_TOKEN_PRIVATE_KEY"))
		assert.Equal(t, "new-token", decryptedToken1)
		assert.Equal(t, "new-token", decryptedToken2)
	})

	t.Run("Should not update account tokens - token is required", func(t *testing.T) {
		// Test
		payload, _ := json.Marshal(AccountTokensRequest{
			DomainId:     "123-controller-update-account-tokens",
			Environments: []string{"default"},
			Token:        "",
		})

		req, _ := http.NewRequest(http.MethodPut, accountController.routeAccountPath+"/123-controller-update-account-tokens", bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.Equal(t, "{\"error\":\"Token is required\"}", response.Body.String())
	})

	t.Run("Should not update account tokens - invalid request", func(t *testing.T) {
		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodPut, accountController.routeAccountPath+"/invalid", bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.Equal(t, "{\"error\":\"Invalid request\"}", response.Body.String())
	})

	t.Run("Should not update account tokens - not found", func(t *testing.T) {
		// Test
		payload, _ := json.Marshal(AccountTokensRequest{
			DomainId:     "not-found",
			Environments: []string{"default"},
			Token:        "new-token",
		})

		req, _ := http.NewRequest(http.MethodPut, accountController.routeAccountPath+"/not-found", bytes.NewBuffer(payload))
		response := executeRequest(req, r, token)

		// Assert
		assert.Equal(t, http.StatusNotFound, response.Code)
		assert.Equal(t, "{\"error\":\"Error fetching account\"}", response.Body.String())
	})

}

func TestDeleteAccountHandler(t *testing.T) {
	token := generateToken("test", time.Minute)

	t.Run("Should delete an account by domain ID / environment", func(t *testing.T) {
		// Create an account
		accountV1.Domain.ID = "123-controller-delete-account"
		accountController.CreateAccountHandler(givenAccountRequest(accountV1))

		// Test
		req, _ := http.NewRequest(http.MethodDelete, accountController.routeAccountPath+"/"+accountV1.Domain.ID+"/"+accountV1.Environment, nil)
		response := executeRequest(req, r, token)

		// Assert
		assert.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("Should not delete an account by domain ID / environment - not found", func(t *testing.T) {
		// Test
		req, _ := http.NewRequest(http.MethodDelete, accountController.routeAccountPath+"/not-found/default", nil)
		response := executeRequest(req, r, token)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Equal(t, "{\"error\":\"Error deleting account: account not found\"}", response.Body.String())
	})
}

func TestUnnauthorizedAccountHandler(t *testing.T) {
	t.Run("Should not create an account - unauthorized", func(t *testing.T) {
		// Test
		payload, _ := json.Marshal(accountV1)
		req, _ := http.NewRequest(http.MethodPost, accountController.routeAccountPath, bytes.NewBuffer(payload))
		response := executeRequest(req, r, "")

		// Assert
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Equal(t, "{\"error\":\"Invalid token\"}", response.Body.String())
	})

	t.Run("Should not fetch an account by domain ID / environment - unauthorized", func(t *testing.T) {
		// Test
		payload := []byte("")
		req, _ := http.NewRequest(http.MethodGet, accountController.routeAccountPath+"/123/default", bytes.NewBuffer(payload))
		response := executeRequest(req, r, "")

		// Assert
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Equal(t, "{\"error\":\"Invalid token\"}", response.Body.String())
	})

	t.Run("Should not update an account - unauthorized", func(t *testing.T) {
		// Test
		payload, _ := json.Marshal(accountV1)
		req, _ := http.NewRequest(http.MethodPut, accountController.routeAccountPath, bytes.NewBuffer(payload))
		response := executeRequest(req, r, "")

		// Assert
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Equal(t, "{\"error\":\"Invalid token\"}", response.Body.String())
	})

	t.Run("Should not delete an account by domain ID / environment - unauthorized", func(t *testing.T) {
		// Test
		req, _ := http.NewRequest(http.MethodDelete, accountController.routeAccountPath+"/123/default", nil)
		response := executeRequest(req, r, "")

		// Assert
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Equal(t, "{\"error\":\"Invalid token\"}", response.Body.String())
	})
}

// Helpers

func givenAccountRequest(data model.Account) (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, accountController.routeAccountPath, nil)

	// Encode the account request as JSON
	body, _ := json.Marshal(data)
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	return w, r
}
