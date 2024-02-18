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
	// Create a sample account request
	accountRequest := model.Account{
		Repository: "switcherapi/switcher-gitops",
		Branch:     "master",
		Domain: model.DomainDetails{
			ID:         "123",
			Name:       "Switcher GitOps",
			Version:    "123",
			LastCommit: "123",
			Status:     "active",
			Message:    "Synced successfully",
		},
		Settings: model.Settings{
			Active:     true,
			Window:     "10m",
			ForcePrune: false,
		},
	}

	// Create a request and response recorder
	w, r := givenAccountRequest(accountRequest)

	// Test
	accountController.CreateAccountHandler(w, r)

	// Assert
	var accountResponse model.Account
	err := json.NewDecoder(w.Body).Decode(&accountResponse)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Nil(t, err)
	assert.Equal(t, accountRequest.Repository, accountResponse.Repository)
}

func givenAccountRequest(data model.Account) (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/account", nil)

	// Encode the account request as JSON
	body, _ := json.Marshal(data)
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	return w, r
}
