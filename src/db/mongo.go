package db

import (
	"context"
	"time"

	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitDb() *mongo.Database {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.GetEnv("MONGO_URI")))

	if err != nil {
		panic(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		utils.LogError("Error connecting to MongoDB: %s", err.Error())
	}

	utils.LogInfo("Connected to MongoDB!")
	return client.Database(config.GetEnv("MONGO_DB"))
}
