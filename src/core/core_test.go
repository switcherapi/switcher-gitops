package core

import (
	"context"
	"os"
	"strings"
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
	gitService := NewGitService("repoURL", "token", "main")
	comparatorService := NewComparatorService()
	coreHandler = NewCoreHandler(accountRepository, gitService, comparatorService)
}

func shutdown() {
	mongoDb.Drop(context.Background())
	mongoDb.Client().Disconnect(context.Background())
}

// Helpers

func canRunIntegratedTests() bool {
	return config.GetEnv("GIT_REPO_URL") != "" &&
		config.GetEnv("GIT_TOKEN") != "" &&
		config.GetEnv("GIT_BRANCH") != "" &&
		config.GetEnv("API_DOMAIN_ID") != ""
}

func AssertNotNil(t *testing.T, object interface{}) {
	if object == nil {
		t.Errorf("Object is nil")
	}
}

func AssertContains(t *testing.T, actual string, expected string) {
	if !strings.Contains(actual, expected) {
		t.Errorf("Expected %v to contain %v", actual, expected)
	}
}

// Fixtures

func givenAccount() model.Account {
	return model.Account{
		Repository:  "switcherapi/switcher-gitops",
		Branch:      "master",
		Environment: "default",
		Domain: model.DomainDetails{
			ID:         "123",
			Name:       "Switcher GitOps",
			Version:    "",
			LastCommit: "",
			LastDate:   "",
			Status:     "",
			Message:    "",
		},
		Settings: model.Settings{
			Active:     true,
			Window:     "5s",
			ForcePrune: false,
		},
	}
}
