package server

import (
	"context"
	"crypto/tls"
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
	if config.GetEnv("SSL_ENABLED") == "true" {
		app.httpServer = startServerWithSsl(app.routerHandlers)
	} else {
		app.httpServer = startServer(app.routerHandlers)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return app.httpServer.Shutdown(ctx)
}

func startServer(routerHandlers *mux.Router) *http.Server {
	port := config.GetEnv("PORT")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: routerHandlers,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			utils.LogError("Failed to listen and serve: %s", err.Error())
			os.Exit(1)
		}
	}()

	utils.LogInfo("[no-SSL] Server started on port %s", port)

	return server
}

func startServerWithSsl(routerHandlers *mux.Router) *http.Server {
	port := config.GetEnv("PORT")
	certFile := config.GetEnv("SSL_CERT_FILE")
	keyFile := config.GetEnv("SSL_KEY_FILE")

	server := &http.Server{
		Addr:    ":" + port,
		Handler: routerHandlers,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	go func() {
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
			utils.LogError("Failed to listen and serve: %s", err.Error())
			os.Exit(1)
		}
	}()

	utils.LogInfo("[SSL] Server started on port %s", port)

	return server
}

func initRoutes(db *mongo.Database, coreHandler *core.CoreHandler) *mux.Router {
	r := mux.NewRouter()

	apiController := controller.NewApiController(coreHandler)
	apiController.RegisterRoutes(r)

	if db != nil {
		accountRepository := repository.NewAccountRepositoryMongo(db)
		accountController := controller.NewAccountController(accountRepository, coreHandler)
		accountController.RegisterRoutes(r)
	}

	return r
}

func initCoreHandler(db *mongo.Database) *core.CoreHandler {
	var coreHandler *core.CoreHandler
	if db == nil {
		return core.NewCoreHandler(nil, nil, nil)
	}

	accountRepository := repository.NewAccountRepositoryMongo(db)
	comparatorService := core.NewComparatorService()
	apiService := core.NewApiService(
		config.GetEnv("SWITCHER_API_JWT_SECRET"),
		config.GetEnv("SWITCHER_API_URL"),
		config.GetEnv("SWITCHER_API_CA_CERT"),
	)

	coreHandler = core.NewCoreHandler(accountRepository, apiService, comparatorService)
	coreHandler.InitCoreHandlerGoroutine()

	return coreHandler
}
