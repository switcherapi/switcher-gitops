package repository

import (
	"context"
	"time"

	"github.com/switcherapi/switcher-gitops/src/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AccountRepository interface {
	Create(account *model.Account) (*model.Account, error)
}

type AccountRepositoryMongo struct {
	Db *mongo.Database
}

func (repo *AccountRepositoryMongo) Create(account *model.Account) (*model.Account, error) {
	collection := repo.Db.Collection(model.CollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, account)

	if err != nil {
		return nil, err
	}

	account.ID = result.InsertedID.(primitive.ObjectID)
	return account, nil
}
