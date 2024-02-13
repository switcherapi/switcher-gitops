package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/controller"
	"github.com/switcherapi/switcher-gitops/src/repository"
)

type App struct {
	httpServer     *http.Server
	routerHandlers *mux.Router
}

func NewApp() *App {
	db := initDb()
	routes := initRoutes(db)

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
			log.Fatalf("Failed to listen and serve: %+v", err)
		}
	}()

	log.Println("Server started on port " + port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return app.httpServer.Shutdown(ctx)
}

func initDb() *mongo.Database {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.GetEnv("MONGO_URI")))

	if err != nil {
		panic(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")
	return client.Database(config.GetEnv("MONGO_DB"))
}

func initRoutes(db *mongo.Database) *mux.Router {
	ApiController := controller.ApiController{}
	AccountController := controller.AccountController{AccountRepository: &repository.AccountRepositoryMongo{Db: db}}

	r := mux.NewRouter()
	r.HandleFunc("/api/check", ApiController.CheckApiHandler).Methods("GET")
	r.HandleFunc("/accounts", AccountController.CreateAccountHandler).Methods("POST")

	return r
}
