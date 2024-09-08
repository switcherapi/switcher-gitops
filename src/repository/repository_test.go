package repository

import (
	"context"
	"os"
	"testing"

	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/db"
	"github.com/switcherapi/switcher-gitops/src/model"
	"go.mongodb.org/mongo-driver/mongo"
)

var mongoDb *mongo.Database
var accountRepository *AccountRepositoryMongo

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
	mongoDb.Drop(context.Background())

	accountRepository = NewAccountRepositoryMongo(mongoDb)
}

func shutdown() {
	mongoDb.Drop(context.Background())
	mongoDb.Client().Disconnect(context.Background())
}

// Fixtures

func givenAccount(active bool) model.Account {
	return model.Account{
		Repository:  "switcherapi/switcher-gitops",
		Branch:      "master",
		Environment: "default",
		Domain: model.DomainDetails{
			ID:         "123-repository-test",
			Name:       "Switcher GitOps",
			Version:    123,
			LastCommit: "123",
			Status:     model.StatusSynced,
			Message:    "Synced successfully",
		},
		Settings: model.Settings{
			Active:     active,
			Window:     "10m",
			ForcePrune: false,
		},
	}
}
