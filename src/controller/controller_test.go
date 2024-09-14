package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/core"
	"github.com/switcherapi/switcher-gitops/src/db"
	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

var mongoDb *mongo.Database
var accountController *AccountController
var apiController *ApiController
var r *mux.Router

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

	// Init modules
	accountRepository := repository.NewAccountRepositoryMongo(mongoDb)
	coreHandler := core.NewCoreHandler(accountRepository, nil, nil)

	apiController = NewApiController(coreHandler)
	accountController = NewAccountController(accountRepository, coreHandler)

	r = mux.NewRouter()
	apiController.RegisterRoutes(r)
	accountController.RegisterRoutes(r)
}

func shutdown() {
	mongoDb.Drop(context.Background())
	mongoDb.Client().Disconnect(context.Background())
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	return rr
}

// Fixtures

var accountV1 = model.Account{
	Repository:  "switcherapi/switcher-gitops",
	Branch:      "master",
	Token:       "github_pat_123",
	Environment: "default",
	Domain: model.DomainDetails{
		ID:         "123-controller-test",
		Name:       "Switcher GitOps",
		Version:    123,
		LastCommit: "123",
		LastDate:   "2021-01-01",
		Status:     model.StatusSynced,
		Message:    "Synced successfully",
	},
	Settings: model.Settings{
		Active:     true,
		Window:     "10m",
		ForcePrune: false,
	},
}

var accountV2 = model.Account{
	Repository:  "switcherapi/switcher-gitops",
	Branch:      "main",
	Environment: "default",
	Domain: model.DomainDetails{
		ID:   "123-controller-test",
		Name: "Switcher GitOps",
	},
	Settings: model.Settings{
		Active:     false,
		Window:     "5m",
		ForcePrune: true,
	},
}
