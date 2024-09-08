package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/controller"
	"github.com/switcherapi/switcher-gitops/src/core"
	"github.com/switcherapi/switcher-gitops/src/db"
	"github.com/switcherapi/switcher-gitops/src/repository"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

type App struct {
	httpServer     *http.Server
	routerHandlers *mux.Router
}

func NewApp() *App {
	db := db.InitDb()
	coreHandler := initCoreHandler(db)
	routes := initRoutes(db, coreHandler)

	return &App{
		routerHandlers: routes,
	}
}

func (app *App) Start() error {
	port := config.GetEnv("PORT")
	app.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: app.routerHandlers,
	}

	go func() {
		if err := app.httpServer.ListenAndServe(); err != nil {
			utils.Log(utils.LogLevelError, "Failed to listen and serve: %s", err.Error())
			os.Exit(1)
		}
	}()

	utils.Log(utils.LogLevelInfo, "Server started on port %s", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return app.httpServer.Shutdown(ctx)
}

func initRoutes(db *mongo.Database, coreHandler *core.CoreHandler) *mux.Router {
	accountRepository := repository.NewAccountRepositoryMongo(db)

	apiController := controller.NewApiController(coreHandler)
	accountController := controller.NewAccountController(accountRepository, coreHandler)

	r := mux.NewRouter()
	apiController.RegisterRoutes(r)
	accountController.RegisterRoutes(r)

	return r
}

func initCoreHandler(db *mongo.Database) *core.CoreHandler {
	accountRepository := repository.NewAccountRepositoryMongo(db)
	comparatorService := core.NewComparatorService()
	apiService := core.NewApiService(
		config.GetEnv("SWITCHER_API_JWT_SECRET"),
		config.GetEnv("SWITCHER_API_URL"),
	)

	coreHandler := core.NewCoreHandler(accountRepository, apiService, comparatorService)
	coreHandler.InitCoreHandlerCoroutine()

	return coreHandler
}
