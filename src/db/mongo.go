package db

import (
	"context"
	"time"

	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/utils"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitDb() *mongo.Database {
	var err error

	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(config.GetEnv("MONGO_URI")))

	if err != nil {
		panic(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		utils.LogError("Error connecting to MongoDB: %s", err.Error())
		return nil
	}

	utils.LogInfo("Connected to MongoDB!")
	return client.Database(config.GetEnv("MONGO_DB"))
}
