package core

import (
	"context"
	"os"
	"testing"

	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/db"
	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

var mongoDb *mongo.Database
var coreHandler *CoreHandler

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

	accountRepository := repository.NewAccountRepositoryMongo(mongoDb)
	apiService := NewApiService("apiKey", "", "")
	comparatorService := NewComparatorService()
	coreHandler = NewCoreHandler(accountRepository, apiService, comparatorService)
}

func shutdown() {
	mongoDb.Drop(context.Background())
	mongoDb.Client().Disconnect(context.Background())
}

// Helpers

func canRunIntegratedTests() bool {
	return config.GetEnv("GIT_REPO_URL") != "" &&
		config.GetEnv("GIT_TOKEN") != "" &&
		config.GetEnv("GIT_TOKEN_READ_ONLY") != "" &&
		config.GetEnv("GIT_BRANCH") != ""
}

// Fixtures

func givenAccount() model.Account {
	return model.Account{
		Repository:  "https://github.com/switcherapi/switcher-gitops-fixture",
		Branch:      "master",
		Token:       "token",
		Environment: "default",
		Domain: model.DomainDetails{
			ID:         "123-core-test",
			Name:       "Switcher GitOps",
			Version:    0,
			LastCommit: "",
			LastDate:   "",
			Status:     model.StatusOutSync,
			Message:    "",
		},
		Settings: &model.Settings{
			Active:     true,
			Window:     "5s",
			ForcePrune: false,
		},
	}
}
