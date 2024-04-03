package repository

import (
	"context"
	"log"
	"time"

	"github.com/switcherapi/switcher-gitops/src/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AccountRepository interface {
	Create(account *model.Account) (*model.Account, error)
	FetchByDomainId(domainId string) (*model.Account, error)
	FetchAllActiveAccounts() ([]model.Account, error)
	Update(account *model.Account) (*model.Account, error)
	DeleteByDomainId(domainId string) error
}

type AccountRepositoryMongo struct {
	Db *mongo.Database
}

type ErrAccountNotFound struct {
	DomainId string
}

func (err ErrAccountNotFound) Error() string {
	return "Account not found for domain.id: " + err.DomainId
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
	update := primitive.M{
		"$set": account,
	}

	result := collection.FindOneAndUpdate(ctx, filter, update)
	if result.Err() != nil {
		return nil, result.Err()
	}

	return account, nil
}

func (repo *AccountRepositoryMongo) DeleteByDomainId(domainId string) error {
	collection, ctx, cancel := getDbContext(repo)
	defer cancel()

	filter := primitive.M{domainIdFilter: domainId}
	result, err := collection.DeleteOne(ctx, filter)

	if result.DeletedCount == 0 {
		return ErrAccountNotFound{DomainId: domainId}
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
		log.Fatal(err)
	}
}
