package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/db"
	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

var mongoDb *mongo.Database

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() {
	os.Setenv("GO_ENV", "test")
	config.InitEnv()
	mongoDb = db.InitDb()
}

func shutdown() {
	// Drop the database after all tests have run
	mongoDb.Drop(context.Background())
	mongoDb.Client().Disconnect(context.Background())
}

func TestCreateAccountHandler(t *testing.T) {
	controller := AccountController{
		AccountRepository: &repository.AccountRepositoryMongo{Db: mongoDb},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/account", nil)

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

	// Encode the account request as JSON
	body, _ := json.Marshal(accountRequest)
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	controller.CreateAccountHandler(w, r)

	// Assert the response status code
	assert.Equal(t, http.StatusCreated, w.Code)

	// Assert the response body
	var accountResponse model.Account
	err := json.NewDecoder(w.Body).Decode(&accountResponse)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, accountRequest.Repository, accountResponse.Repository)
}
