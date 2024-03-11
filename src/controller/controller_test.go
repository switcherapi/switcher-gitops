package controller

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
var accountController *AccountController
var apiController *ApiController

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

	apiController = NewApiController()
	accountController = NewAccountController(accountRepository)
}

func shutdown() {
	mongoDb.Drop(context.Background())
	mongoDb.Client().Disconnect(context.Background())
}

// Fixtures

var accountV1 = model.Account{
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

var accountV2 = model.Account{
	Repository: "switcherapi/switcher-gitops",
	Branch:     "master",
	Domain: model.DomainDetails{
		ID:         "123",
		Name:       "Switcher GitOps",
		Version:    "123",
		LastCommit: "123",
		Status:     "active",
	},
	Settings: model.Settings{
		Active:     false,
		Window:     "5m",
		ForcePrune: true,
	},
}
