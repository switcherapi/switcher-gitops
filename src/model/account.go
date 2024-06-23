package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	CollectionName = "accounts"
)

const (
	StatusCreated = "Created"
	StatusSynced  = "Synced"
	StatusOutSync = "OutSync"
	StatusError   = "Error"
)

type Account struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Repository  string             `json:"repository"`
	Branch      string             `json:"branch"`
	Environment string             `json:"environment"`
	Domain      DomainDetails      `json:"domain"`
	Settings    Settings           `json:"settings"`
}

type DomainDetails struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	LastCommit string `json:"lastcommit"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

type Settings struct {
	Active     bool   `json:"active"`
	Window     string `json:"window"`
	ForcePrune bool   `json:"forceprune"`
}
