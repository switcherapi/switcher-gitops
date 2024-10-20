package repository

import (
	"context"
	"errors"
	"time"

	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AccountRepository interface {
	Create(account *model.Account) (*model.Account, error)
	FetchByAccountId(accountId string) (*model.Account, error)
	FetchByDomainIdEnvironment(domainId string, environment string) (*model.Account, error)
	FetchAllByDomainId(domainId string) []model.Account
	FetchAllAccounts() []model.Account
	UpdateByDomainEnvironment(account *model.Account) (*model.Account, error)
	DeleteByAccountId(accountId string) error
	DeleteByDomainIdEnvironment(domainId string, environment string) error
}

type AccountRepositoryMongo struct {
	Db *mongo.Database
}

const domainIdFilter = "domain.id"
const environmentFilter = "environment"

func NewAccountRepositoryMongo(db *mongo.Database) *AccountRepositoryMongo {
	registerAccountRepositoryValidators(db)
	return &AccountRepositoryMongo{Db: db}
}

func (repo *AccountRepositoryMongo) Create(account *model.Account) (*model.Account, error) {
	collection, ctx, cancel := getDbContext(repo)
	defer cancel()

	result, err := collection.InsertOne(ctx, account)
	if err != nil {
		return nil, err
	}

	account.ID = result.InsertedID.(primitive.ObjectID)
	return account, nil
}

func (repo *AccountRepositoryMongo) UpdateByDomainEnvironment(account *model.Account) (*model.Account, error) {
	collection, ctx, cancel := getDbContext(repo)
	defer cancel()

	filter := primitive.M{domainIdFilter: account.Domain.ID, environmentFilter: account.Environment}
	update := getUpdateFields(account)

	var updatedAccount model.Account
	err := collection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().
		SetReturnDocument(options.After)).
		Decode(&updatedAccount)

	if err != nil {
		return nil, err
	}

	return &updatedAccount, nil
}

func (repo *AccountRepositoryMongo) FetchByAccountId(accountId string) (*model.Account, error) {
	objectId, _ := primitive.ObjectIDFromHex(accountId)
	filter := bson.M{"_id": objectId}
	return repo.fetchOne(filter)
}

func (repo *AccountRepositoryMongo) FetchByDomainIdEnvironment(domainId string, environment string) (*model.Account, error) {
	filter := primitive.M{domainIdFilter: domainId, environmentFilter: environment}
	return repo.fetchOne(filter)
}

func (repo *AccountRepositoryMongo) FetchAllByDomainId(domainId string) []model.Account {
	filter := primitive.M{domainIdFilter: domainId}
	return repo.fetchMany(filter)
}

func (repo *AccountRepositoryMongo) FetchAllAccounts() []model.Account {
	filter := bson.M{}
	return repo.fetchMany(filter)
}

func (repo *AccountRepositoryMongo) DeleteByAccountId(accountId string) error {
	objectId, _ := primitive.ObjectIDFromHex(accountId)
	filter := bson.M{"_id": objectId}
	return repo.deleteOne(filter)
}

func (repo *AccountRepositoryMongo) DeleteByDomainIdEnvironment(domainId string, environment string) error {
	filter := primitive.M{domainIdFilter: domainId, environmentFilter: environment}
	return repo.deleteOne(filter)
}

func (repo *AccountRepositoryMongo) deleteOne(filter primitive.M) error {
	collection, ctx, cancel := getDbContext(repo)
	defer cancel()

	result, err := collection.DeleteOne(ctx, filter)
	if result.DeletedCount == 0 {
		return errors.New("account not found")
	}

	return err
}

func (repo *AccountRepositoryMongo) fetchMany(filter primitive.M) []model.Account {
	collection, ctx, cancel := getDbContext(repo)
	defer cancel()

	cursor, _ := collection.Find(ctx, filter)

	var accounts []model.Account
	for cursor.Next(ctx) {
		var account model.Account
		err := cursor.Decode(&account)
		if err == nil {
			accounts = append(accounts, account)
		}
	}

	return accounts
}

func (repo *AccountRepositoryMongo) fetchOne(filter primitive.M) (*model.Account, error) {
	collection, ctx, cancel := getDbContext(repo)
	defer cancel()

	var account model.Account
	err := collection.FindOne(ctx, filter).Decode(&account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func getDbContext(repo *AccountRepositoryMongo) (*mongo.Collection, context.Context, context.CancelFunc) {
	collection := repo.Db.Collection(model.CollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	return collection, ctx, cancel
}

func registerAccountRepositoryValidators(db *mongo.Database) {
	collection := db.Collection(model.CollectionName)
	indexOptions := options.Index().SetUnique(true)
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: domainIdFilter, Value: 1},
			{Key: environmentFilter, Value: 1},
		},
		Options: indexOptions,
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		utils.LogError("Error creating index for account (environment, domain.id): %s", err.Error())
	}
}

func getUpdateFields(account *model.Account) bson.M {
	setMap := bson.M{}

	setIfNotEmpty(setMap, "environment", account.Environment)
	setIfNotEmpty(setMap, "repository", account.Repository)
	setIfNotEmpty(setMap, "branch", account.Branch)
	setIfNotEmpty(setMap, "path", account.Path)
	setIfNotEmpty(setMap, "token", account.Token)

	setIfNotEmpty(setMap, "domain.name", account.Domain.Name)
	setIfNotEmpty(setMap, "domain.version", account.Domain.Version)
	setIfNotEmpty(setMap, "domain.lastCommit", account.Domain.LastCommit)
	setIfNotEmpty(setMap, "domain.lastDate", account.Domain.LastDate)
	setIfNotEmpty(setMap, "domain.status", account.Domain.Status)
	setIfNotEmpty(setMap, "domain.message", account.Domain.Message)

	setIfNotEmpty(setMap, "settings.window", account.Settings.Window)

	setMap["settings.active"] = account.Settings.Active
	setMap["settings.forcePrune"] = account.Settings.ForcePrune

	update := bson.M{"$set": setMap}
	return update
}

func setIfNotEmpty(setMap bson.M, key string, value interface{}) {
	switch v := value.(type) {
	case string:
		if v != "" {
			setMap[key] = v
		}
	case int:
		if v != 0 {
			setMap[key] = v
		}
	}
}
