package controller

import (
	"context"
	"os"
	"testing"

	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/db"
	"github.com/switcherapi/switcher-gitops/src/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

var mongoDb *mongo.Database
var accountController AccountController
var apiController ApiController

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

	apiController = ApiController{}
	accountController = AccountController{
		AccountRepository: &repository.AccountRepositoryMongo{Db: mongoDb},
	}
}

func shutdown() {
	// Drop the database after all tests have run
	mongoDb.Drop(context.Background())
	mongoDb.Client().Disconnect(context.Background())
}
