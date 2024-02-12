package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/switcherapi/switcher-gitops/src/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AccountController struct {
	Db  mongo.Database
	Ctx context.Context
}

func (controller *AccountController) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var account model.Account
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	collection := controller.Db.Collection(model.CollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, _ := collection.InsertOne(ctx, account)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	account.ID = result.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(account)
}
