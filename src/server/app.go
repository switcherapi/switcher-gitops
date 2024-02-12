package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/controller"
)

type App struct {
	httpServer     *http.Server
	routerHandlers *mux.Router
}

func NewApp() *App {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := initDb(ctx)
	routes := initRoutes(db, ctx)

	return &App{
		routerHandlers: routes,
	}
}

func (app *App) Start() {
	port := config.GetEnv("PORT")
	app.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: app.routerHandlers,
	}

	log.Println("Server started on port " + port)
	app.httpServer.ListenAndServe()
}

func initDb(ctx context.Context) *mongo.Database {
	var err error

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.GetEnv("MONGO_URI")))
	db := client.Database(config.GetEnv("MONGO_DB"))

	if err != nil {
		panic(err)
	}

	log.Println("Connected to MongoDB!")
	return db
}

func initRoutes(db *mongo.Database, ctx context.Context) *mux.Router {
	AccountController := controller.AccountController{Db: *db, Ctx: ctx}

	r := mux.NewRouter()
	r.HandleFunc("/accounts", AccountController.CreateAccountHandler).Methods("POST")

	return r
}
