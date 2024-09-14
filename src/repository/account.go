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
	FetchByDomainId(domainId string) (*model.Account, error)
	FetchAllActiveAccounts() ([]model.Account, error)
	Update(account *model.Account) (*model.Account, error)
	DeleteByAccountId(accountId string) error
	DeleteByDomainId(domainId string) error
}

type AccountRepositoryMongo struct {
	Db *mongo.Database
}

const domainIdFilter = "domain.id"
const settingsActiveFilter = "settings.active"

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

func (repo *AccountRepositoryMongo) FetchByAccountId(accountId string) (*model.Account, error) {
	collection, ctx, cancel := getDbContext(repo)
	defer cancel()

	var account model.Account
	objectId, _ := primitive.ObjectIDFromHex(accountId)
	filter := bson.M{"_id": objectId}
	err := collection.FindOne(ctx, filter).Decode(&account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (repo *AccountRepositoryMongo) FetchByDomainId(domainId string) (*model.Account, error) {
	collection, ctx, cancel := getDbContext(repo)
	defer cancel()

	var account model.Account
	filter := primitive.M{domainIdFilter: domainId}
	err := collection.FindOne(ctx, filter).Decode(&account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (repo *AccountRepositoryMongo) FetchAllActiveAccounts() ([]model.Account, error) {
	collection, ctx, cancel := getDbContext(repo)
	defer cancel()

	filter := primitive.M{settingsActiveFilter: true}
	cursor, _ := collection.Find(ctx, filter)

	var accounts []model.Account
	for cursor.Next(ctx) {
		var account model.Account
		err := cursor.Decode(&account)
		if err == nil {
			accounts = append(accounts, account)
		}
	}

	return accounts, nil
}

func (repo *AccountRepositoryMongo) Update(account *model.Account) (*model.Account, error) {
	collection, ctx, cancel := getDbContext(repo)
	defer cancel()

	filter := primitive.M{domainIdFilter: account.Domain.ID}
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

func (repo *AccountRepositoryMongo) DeleteByAccountId(accountId string) error {
	collection, ctx, cancel := getDbContext(repo)
	defer cancel()

	objectId, _ := primitive.ObjectIDFromHex(accountId)
	filter := bson.M{"_id": objectId}
	result, err := collection.DeleteOne(ctx, filter)

	if result.DeletedCount == 0 {
		return errors.New("Account not found for id: " + accountId)
	}

	return err
}

func (repo *AccountRepositoryMongo) DeleteByDomainId(domainId string) error {
	collection, ctx, cancel := getDbContext(repo)
	defer cancel()

	filter := primitive.M{domainIdFilter: domainId}
	result, err := collection.DeleteOne(ctx, filter)

	if result.DeletedCount == 0 {
		return errors.New("Account not found for domain.id: " + domainId)
	}

	return err
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
		Keys:    bson.M{domainIdFilter: 1},
		Options: indexOptions,
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		utils.Log(utils.LogLevelError, "Error creating index for domain.id: %s", err.Error())
	}
}

func getUpdateFields(account *model.Account) bson.M {
	setMap := bson.M{}

	setMap["repository"] = account.Repository
	setMap["branch"] = account.Branch
	setMap["environment"] = account.Environment
	setMap["domain.name"] = account.Domain.Name
	setMap["settings.active"] = account.Settings.Active
	setMap["settings.window"] = account.Settings.Window
	setMap["settings.forcePrune"] = account.Settings.ForcePrune

	setIfNotEmpty(setMap, "token", account.Token)
	setIfNotEmpty(setMap, "domain.version", account.Domain.Version)
	setIfNotEmpty(setMap, "domain.lastCommit", account.Domain.LastCommit)
	setIfNotEmpty(setMap, "domain.lastDate", account.Domain.LastDate)
	setIfNotEmpty(setMap, "domain.status", account.Domain.Status)
	setIfNotEmpty(setMap, "domain.message", account.Domain.Message)

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
